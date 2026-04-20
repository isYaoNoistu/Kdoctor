package host

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/config"
	"kdoctor/internal/snapshot"
	disktransport "kdoctor/internal/transport/disk"
	dockertransport "kdoctor/internal/transport/docker"
	"kdoctor/internal/transport/tcp"
)

type Collector struct{}

func (Collector) Collect(ctx context.Context, env *config.Runtime, compose *snapshot.ComposeSnapshot, docker *snapshot.DockerSnapshot) *snapshot.HostSnapshot {
	if env == nil || !env.EnableHost {
		return nil
	}

	out := &snapshot.HostSnapshot{
		Raw: map[string]string{},
	}

	diskTargets := collectDiskTargets(env, compose, docker)
	configuredPorts := collectConfiguredPorts(env)
	hostContext := len(diskTargets) > 0 || len(configuredPorts) > 0 || (docker != nil && docker.Available)
	if len(diskTargets) == 0 && !hostContext {
		out.Collected = false
		out.Errors = append(out.Errors, "host-level evidence is not available from the current input mode")
		return out
	}

	for _, target := range diskTargets {
		usage, err := disktransport.Stat(target)
		if err != nil {
			out.Errors = append(out.Errors, err.Error())
			continue
		}
		out.DiskUsages = append(out.DiskUsages, snapshot.DiskUsage{
			Path:            usage.Path,
			TotalBytes:      usage.TotalBytes,
			AvailableBytes:  usage.AvailableBytes,
			UsedBytes:       usage.UsedBytes,
			UsedPercent:     usage.UsedPercent,
			TotalInodes:     usage.TotalInodes,
			AvailableInodes: usage.AvailableInodes,
			UsedInodes:      usage.UsedInodes,
			UsedInodePct:    usage.UsedInodePct,
		})
	}

	if hostContext {
		listeners := append(collectListenerTargets(compose), configuredPorts...)
		for _, address := range dedupeStrings(listeners) {
			result := tcp.Dial(ctx, address, env.TCPTimeout)
			out.PortChecks = append(out.PortChecks, snapshot.EndpointCheck{
				Kind:       "listener",
				Address:    address,
				Reachable:  result.Reachable,
				DurationMs: result.Duration.Milliseconds(),
				Error:      result.Error,
			})
		}
	}

	signals := collectSystemSignals(ctx)
	if signals.FD != nil {
		out.FD = signals.FD
	}
	if containerFD := collectContainerFD(ctx, docker); len(containerFD) > 0 {
		out.ContainerFD = containerFD
	}
	if signals.Memory != nil {
		out.Memory = signals.Memory
	}
	if len(signals.ListenPorts) > 0 {
		out.ObservedListenPorts = append(out.ObservedListenPorts, signals.ListenPorts...)
	}
	out.Errors = append(out.Errors, signals.Errors...)

	out.Collected = len(diskTargets) > 0 || hostContext || len(out.Errors) > 0
	out.Available = len(out.DiskUsages) > 0 || len(out.PortChecks) > 0 || out.FD != nil || len(out.ContainerFD) > 0 || out.Memory != nil || len(out.ObservedListenPorts) > 0
	return out
}

func collectContainerFD(ctx context.Context, docker *snapshot.DockerSnapshot) []snapshot.ContainerFDStat {
	if docker == nil || !docker.Available {
		return nil
	}
	stats := make([]snapshot.ContainerFDStat, 0, len(docker.Containers))
	for _, container := range docker.Containers {
		if !container.Running || strings.TrimSpace(container.Name) == "" {
			continue
		}
		item := snapshot.ContainerFDStat{Name: container.Name}
		soft, hard, err := dockertransport.ProcessOpenFileLimit(ctx, container.Name)
		if err != nil {
			item.Error = err.Error()
			stats = append(stats, item)
			continue
		}
		item.SoftLimit = soft
		item.HardLimit = hard
		stats = append(stats, item)
	}
	return stats
}

func collectDiskTargets(env *config.Runtime, compose *snapshot.ComposeSnapshot, docker *snapshot.DockerSnapshot) []string {
	targets := []string{}
	if path := existingPath(strings.TrimSpace(env.LogDir)); path != "" {
		targets = append(targets, path)
	}
	for _, diskPath := range env.Config.Host.DiskPaths {
		if path := existingPath(strings.TrimSpace(diskPath)); path != "" {
			targets = append(targets, path)
		}
	}

	for _, service := range composeutil.KafkaServices(compose) {
		containerName := service.ContainerName
		if strings.TrimSpace(containerName) == "" {
			containerName = service.ServiceName
		}
		for _, logDir := range containerLogDirs(service) {
			if path := resolveContainerPath(compose, docker, containerName, service, logDir); path != "" {
				targets = append(targets, path)
			}
		}
	}

	return dedupePaths(targets)
}

func collectListenerTargets(compose *snapshot.ComposeSnapshot) []string {
	targets := []string{}
	for _, service := range composeutil.KafkaServices(compose) {
		listeners, err := composeutil.ParseListeners(service.Environment["KAFKA_CFG_LISTENERS"])
		if err != nil {
			continue
		}
		for _, listener := range listeners {
			host := strings.TrimSpace(listener.Host)
			switch host {
			case "", "0.0.0.0", "::":
				host = "127.0.0.1"
			}
			targets = append(targets, fmt.Sprintf("%s:%d", host, listener.Port))
		}
	}
	return dedupeStrings(targets)
}

func collectConfiguredPorts(env *config.Runtime) []string {
	if env == nil {
		return nil
	}
	addresses := []string{}
	for _, port := range env.Config.Host.CheckPorts {
		if port <= 0 {
			continue
		}
		addresses = append(addresses, fmt.Sprintf("127.0.0.1:%d", port))
	}
	return dedupeStrings(addresses)
}

func containerLogDirs(service composeutil.KafkaService) []string {
	dirs := composeutil.ParseCSV(service.Environment["KAFKA_CFG_LOG_DIRS"])
	if metadata := strings.TrimSpace(service.Environment["KAFKA_CFG_METADATA_LOG_DIR"]); metadata != "" {
		dirs = append(dirs, metadata)
	}
	return dedupeStrings(dirs)
}

func resolveContainerPath(compose *snapshot.ComposeSnapshot, docker *snapshot.DockerSnapshot, containerName string, service composeutil.KafkaService, containerPath string) string {
	if path, root := mapFromDocker(docker, containerName, containerPath); path != "" {
		return existingWithinRoot(path, root)
	}
	if compose != nil {
		if path, ok := composeutil.MapContainerPathToHost(compose.SourcePath, service, containerPath); ok {
			if root := composeRootForContainerPath(compose.SourcePath, service, containerPath); root != "" {
				return existingWithinRoot(path, root)
			}
			return existingPath(path)
		}
	}
	return ""
}

func mapFromDocker(docker *snapshot.DockerSnapshot, containerName string, containerPath string) (string, string) {
	if docker == nil {
		return "", ""
	}
	for _, container := range docker.Containers {
		if container.Name != containerName {
			continue
		}
		bestDestination := ""
		bestSource := ""
		for _, mount := range container.Mounts {
			if !hasContainerPathPrefix(containerPath, mount.Destination) {
				continue
			}
			if len(mount.Destination) > len(bestDestination) {
				bestDestination = mount.Destination
				bestSource = mount.Source
			}
		}
		if bestDestination == "" || strings.TrimSpace(bestSource) == "" {
			return "", ""
		}

		suffix := strings.TrimPrefix(containerPath, bestDestination)
		suffix = strings.TrimPrefix(suffix, "/")
		if suffix == "" {
			return filepath.Clean(bestSource), filepath.Clean(bestSource)
		}
		suffix = strings.ReplaceAll(suffix, "/", string(filepath.Separator))
		root := filepath.Clean(bestSource)
		return filepath.Clean(filepath.Join(bestSource, suffix)), root
	}
	return "", ""
}

func composeRootForContainerPath(composePath string, service composeutil.KafkaService, containerPath string) string {
	for _, volume := range service.Volumes {
		mount := composeutil.ParseVolumeSpec(volume)
		if mount.Destination == "" || mount.NamedVolume {
			continue
		}
		if !hasContainerPathPrefix(containerPath, mount.Destination) {
			continue
		}
		return composeutil.ResolveHostPath(composePath, mount.Source)
	}
	return ""
}

func existingPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	path = filepath.Clean(path)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func existingWithinRoot(path string, root string) string {
	path = filepath.Clean(strings.TrimSpace(path))
	root = filepath.Clean(strings.TrimSpace(root))
	if path == "" || root == "" {
		return ""
	}
	for {
		if _, err := os.Stat(path); err == nil {
			return path
		}
		if path == root {
			return ""
		}
		parent := filepath.Dir(path)
		if parent == path {
			return ""
		}
		path = parent
	}
}

func dedupePaths(paths []string) []string {
	values := dedupeStrings(paths)
	sort.Strings(values)
	return values
}

func dedupeStrings(values []string) []string {
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
	return out
}

func hasContainerPathPrefix(pathValue, prefix string) bool {
	pathValue = strings.TrimSuffix(strings.TrimSpace(pathValue), "/")
	prefix = strings.TrimSuffix(strings.TrimSpace(prefix), "/")
	if pathValue == prefix {
		return true
	}
	return strings.HasPrefix(pathValue, prefix+"/")
}

type systemSignals struct {
	FD          *snapshot.FDStats
	Memory      *snapshot.MemoryStats
	ListenPorts []int
	Errors      []string
}
