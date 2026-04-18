package storage

import (
	"path/filepath"
	"sort"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/snapshot"
)

func services(bundle *snapshot.Bundle) []composeutil.KafkaService {
	if bundle == nil || bundle.Compose == nil {
		return nil
	}
	return composeutil.KafkaServices(bundle.Compose)
}

func storagePaths(service composeutil.KafkaService) (logDirs []string, metadataDir string) {
	logDirs = composeutil.ParseCSV(service.Environment["KAFKA_CFG_LOG_DIRS"])
	metadataDir = strings.TrimSpace(service.Environment["KAFKA_CFG_METADATA_LOG_DIR"])
	return dedupe(logDirs), metadataDir
}

func roles(service composeutil.KafkaService) []string {
	return composeutil.ParseCSV(service.Environment["KAFKA_CFG_PROCESS_ROLES"])
}

func hasRole(values []string, want string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), want) {
			return true
		}
	}
	return false
}

func overlaps(left string, right string) bool {
	left = cleanContainerPath(left)
	right = cleanContainerPath(right)
	if left == "" || right == "" {
		return false
	}
	if left == right {
		return true
	}
	return strings.HasPrefix(left, right+"/") || strings.HasPrefix(right, left+"/")
}

func coveringVolume(service composeutil.KafkaService, containerPath string) (composeutil.VolumeMount, bool) {
	containerPath = cleanContainerPath(containerPath)
	best := composeutil.VolumeMount{}
	bestLen := -1
	for _, raw := range service.Volumes {
		mount := composeutil.ParseVolumeSpec(raw)
		destination := cleanContainerPath(mount.Destination)
		if destination == "" {
			continue
		}
		if containerPath != destination && !strings.HasPrefix(containerPath, destination+"/") {
			continue
		}
		if len(destination) > bestLen {
			best = mount
			bestLen = len(destination)
		}
	}
	if bestLen == -1 {
		return composeutil.VolumeMount{}, false
	}
	return best, true
}

func cleanContainerPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(value))
}

func dedupe(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = cleanContainerPath(value)
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
