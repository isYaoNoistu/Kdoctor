package lint

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type MetadataDirChecker struct{}

func (MetadataDirChecker) ID() string     { return "CFG-014" }
func (MetadataDirChecker) Name() string   { return "metadata_logdir_planning" }
func (MetadataDirChecker) Module() string { return "lint" }

func (MetadataDirChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	composeServices := composeutil.KafkaServices(getCompose(snap))
	services := kafkaServices(getCompose(snap))
	if len(services) == 0 {
		return rule.NewSkip("CFG-014", "metadata_logdir_planning", "lint", "compose Kafka services not available")
	}

	warnings := 0
	failures := 0
	evidence := []string{}
	for _, service := range services {
		if !contains(service.ProcessRoles, "controller") {
			continue
		}
		if service.MetadataLogDir == "" {
			warnings++
			evidence = append(evidence, fmt.Sprintf("service=%s metadata_log_dir_missing=true", service.ServiceName))
			continue
		}
		evidence = append(evidence, fmt.Sprintf("service=%s metadata_log_dir=%s", service.ServiceName, service.MetadataLogDir))
		if mount, ok := lintCoveringVolume(findService(composeServices, service.ServiceName), service.MetadataLogDir); ok {
			evidence = append(evidence, fmt.Sprintf("service=%s metadata_volume=%s->%s named_volume=%t", service.ServiceName, mount.Source, mount.Destination, mount.NamedVolume))
			if mount.NamedVolume {
				warnings++
			}
		} else {
			failures++
		}
	}

	result := rule.NewPass("CFG-014", "metadata_logdir_planning", "lint", "metadata.log.dir planning is explicit and backed by storage")
	result.Evidence = evidence
	switch {
	case failures > 0:
		result = rule.NewFail("CFG-014", "metadata_logdir_planning", "lint", "some controller metadata directories are not backed by a clear compose volume mapping")
		result.Evidence = evidence
		result.NextActions = []string{"bind-mount metadata.log.dir to a stable host path", "avoid leaving KRaft metadata only in ephemeral container layers", "review controller storage planning together with data directory planning"}
	case warnings > 0:
		result = rule.NewWarn("CFG-014", "metadata_logdir_planning", "lint", "metadata.log.dir is present but still relies on implicit or named-volume planning")
		result.Evidence = evidence
		result.NextActions = []string{"prefer an explicit host path for metadata.log.dir in production-like environments", "confirm named volumes meet your persistence and backup expectations", "keep controller metadata storage planning easy to audit during incidents"}
	}
	return result
}

func findService(services []composeutil.KafkaService, name string) composeutil.KafkaService {
	for _, service := range services {
		if service.ServiceName == name {
			return service
		}
	}
	return composeutil.KafkaService{}
}

func lintCoveringVolume(service composeutil.KafkaService, containerPath string) (composeutil.VolumeMount, bool) {
	containerPath = lintCleanContainerPath(containerPath)
	best := composeutil.VolumeMount{}
	bestLen := -1
	for _, raw := range service.Volumes {
		mount := composeutil.ParseVolumeSpec(raw)
		destination := lintCleanContainerPath(mount.Destination)
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

func lintCleanContainerPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(value))
}
