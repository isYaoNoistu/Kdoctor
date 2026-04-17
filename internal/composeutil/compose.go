package composeutil

import (
	"net"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"kdoctor/internal/snapshot"
)

type KafkaService struct {
	ServiceName   string
	ContainerName string
	Image         string
	NetworkMode   string
	Environment   map[string]string
	Volumes       []string
}

type ListenerEndpoint struct {
	Name string
	Host string
	Port int
	Raw  string
}

type VolumeMount struct {
	Source      string
	Destination string
	Mode        string
	NamedVolume bool
}

func KafkaServices(compose *snapshot.ComposeSnapshot) []KafkaService {
	if compose == nil {
		return nil
	}

	names := make([]string, 0, len(compose.Services))
	for name := range compose.Services {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]KafkaService, 0, len(names))
	for _, name := range names {
		service := compose.Services[name]
		if !IsKafkaService(service) {
			continue
		}
		out = append(out, KafkaService{
			ServiceName:   name,
			ContainerName: service.ContainerName,
			Image:         service.Image,
			NetworkMode:   service.NetworkMode,
			Environment:   cloneMap(service.Environment),
			Volumes:       append([]string(nil), service.Volumes...),
		})
	}
	return out
}

func ContainerNames(compose *snapshot.ComposeSnapshot, explicit []string) []string {
	if len(explicit) > 0 {
		return dedupe(explicit)
	}

	services := KafkaServices(compose)
	names := make([]string, 0, len(services))
	for _, service := range services {
		if strings.TrimSpace(service.ContainerName) != "" {
			names = append(names, service.ContainerName)
			continue
		}
		names = append(names, service.ServiceName)
	}
	return dedupe(names)
}

func IsKafkaService(service snapshot.ComposeService) bool {
	if strings.Contains(strings.ToLower(service.Image), "kafka") {
		return true
	}
	if _, ok := service.Environment["KAFKA_CFG_NODE_ID"]; ok {
		return true
	}
	if _, ok := service.Environment["KAFKA_CFG_PROCESS_ROLES"]; ok {
		return true
	}
	return false
}

func ParseCSV(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func ParseListeners(input string) (map[string]ListenerEndpoint, error) {
	out := map[string]ListenerEndpoint{}
	for _, item := range ParseCSV(input) {
		parts := strings.SplitN(item, "://", 2)
		if len(parts) != 2 {
			return nil, strconv.ErrSyntax
		}
		host, portStr, err := net.SplitHostPort(parts[1])
		if err != nil {
			return nil, err
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
		out[strings.TrimSpace(parts[0])] = ListenerEndpoint{
			Name: strings.TrimSpace(parts[0]),
			Host: host,
			Port: port,
			Raw:  item,
		}
	}
	return out, nil
}

func ParseVoters(input string) (map[int]string, error) {
	out := map[int]string{}
	for _, item := range ParseCSV(input) {
		parts := strings.SplitN(item, "@", 2)
		if len(parts) != 2 {
			return nil, strconv.ErrSyntax
		}
		id, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, err
		}
		out[id] = strings.TrimSpace(parts[1])
	}
	return out, nil
}

func ParseVolumeSpec(spec string) VolumeMount {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return VolumeMount{}
	}

	rest := spec
	mode := ""
	if idx := strings.LastIndex(rest, ":"); idx > 0 {
		tail := strings.TrimSpace(rest[idx+1:])
		if tail == "ro" || tail == "rw" || tail == "z" {
			mode = tail
			rest = rest[:idx]
		}
	}

	idx := strings.LastIndex(rest, ":")
	if idx <= 0 {
		return VolumeMount{}
	}

	source := strings.TrimSpace(rest[:idx])
	destination := strings.TrimSpace(rest[idx+1:])
	return VolumeMount{
		Source:      source,
		Destination: destination,
		Mode:        mode,
		NamedVolume: !looksLikeHostPath(source),
	}
}

func ResolveHostPath(composePath, source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return ""
	}
	if filepath.IsAbs(source) {
		return filepath.Clean(source)
	}

	base := filepath.Dir(composePath)
	if strings.TrimSpace(base) == "" || base == "." {
		base = "."
	}
	return filepath.Clean(filepath.Join(base, source))
}

func MapContainerPathToHost(composePath string, service KafkaService, containerPath string) (string, bool) {
	containerPath = strings.TrimSpace(containerPath)
	if containerPath == "" {
		return "", false
	}

	bestMatch := ""
	bestSource := ""
	for _, volume := range service.Volumes {
		mount := ParseVolumeSpec(volume)
		if mount.Destination == "" || mount.NamedVolume {
			continue
		}
		if !hasPathPrefix(containerPath, mount.Destination) {
			continue
		}
		if len(mount.Destination) > len(bestMatch) {
			bestMatch = mount.Destination
			bestSource = ResolveHostPath(composePath, mount.Source)
		}
	}
	if bestMatch == "" || bestSource == "" {
		return "", false
	}

	suffix := strings.TrimPrefix(containerPath, bestMatch)
	suffix = strings.TrimPrefix(suffix, "/")
	if suffix == "" {
		return filepath.Clean(bestSource), true
	}
	suffix = strings.ReplaceAll(suffix, "/", string(filepath.Separator))
	return filepath.Clean(filepath.Join(bestSource, suffix)), true
}

func cloneMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func dedupe(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func looksLikeHostPath(source string) bool {
	if strings.HasPrefix(source, ".") || strings.HasPrefix(source, "/") || strings.HasPrefix(source, "~") {
		return true
	}
	if len(source) >= 3 && ((source[1] == ':' && source[2] == '\\') || (source[1] == ':' && source[2] == '/')) {
		return true
	}
	return false
}

func hasPathPrefix(pathValue, prefix string) bool {
	pathValue = strings.TrimSuffix(strings.TrimSpace(pathValue), "/")
	prefix = strings.TrimSuffix(strings.TrimSpace(prefix), "/")
	if pathValue == prefix {
		return true
	}
	return strings.HasPrefix(pathValue, prefix+"/")
}
