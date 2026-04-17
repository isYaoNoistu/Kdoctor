package docker

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type MountChecker struct{}

func (MountChecker) ID() string     { return "DKR-004" }
func (MountChecker) Name() string   { return "log_mounts" }
func (MountChecker) Module() string { return "docker" }

func (MountChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	docker := dockerSnap(bundle)
	if docker == nil || !docker.Collected {
		return rule.NewSkip("DKR-004", "log_mounts", "docker", "docker runtime is not enabled in the current input mode")
	}
	if !docker.Available {
		result := rule.NewSkip("DKR-004", "log_mounts", "docker", "docker runtime is not available on the current execution host")
		result.Evidence = append(result.Evidence, docker.Errors...)
		return result
	}

	services := composeutil.KafkaServices(getCompose(bundle))
	if len(services) == 0 {
		return rule.NewSkip("DKR-004", "log_mounts", "docker", "compose Kafka services are not available for mount validation")
	}

	containers := dockerContainerMap(docker)
	missingMounts := 0
	evidence := []string{}
	for _, service := range services {
		containerName := service.ContainerName
		if strings.TrimSpace(containerName) == "" {
			containerName = service.ServiceName
		}
		container, ok := containers[containerName]
		if !ok {
			missingMounts++
			evidence = append(evidence, fmt.Sprintf("%s container not found", containerName))
			continue
		}

		requiredPaths := composeutil.ParseCSV(service.Environment["KAFKA_CFG_LOG_DIRS"])
		if metadata := strings.TrimSpace(service.Environment["KAFKA_CFG_METADATA_LOG_DIR"]); metadata != "" {
			requiredPaths = append(requiredPaths, metadata)
		}
		for _, requiredPath := range requiredPaths {
			if mounted(requiredPath, container.Mounts) {
				evidence = append(evidence, fmt.Sprintf("%s mounted for %s", requiredPath, containerName))
				continue
			}
			missingMounts++
			evidence = append(evidence, fmt.Sprintf("%s not backed by a docker mount in %s", requiredPath, containerName))
		}
	}

	result := rule.NewPass("DKR-004", "log_mounts", "docker", "Kafka data and metadata paths are backed by docker mounts")
	result.Evidence = evidence
	if missingMounts > 0 {
		result = rule.NewFail("DKR-004", "log_mounts", "docker", "some Kafka data or metadata paths are not backed by docker mounts")
		result.Evidence = evidence
		result.NextActions = []string{"bind-mount Kafka data and metadata directories", "verify compose volume declarations", "avoid storing Kafka state only in ephemeral container layers"}
	}
	return result
}

func getCompose(bundle *snapshot.Bundle) *snapshot.ComposeSnapshot {
	if bundle == nil {
		return nil
	}
	return bundle.Compose
}

func mounted(containerPath string, mounts []snapshot.DockerMount) bool {
	containerPath = strings.TrimSuffix(strings.TrimSpace(containerPath), "/")
	for _, mount := range mounts {
		destination := strings.TrimSuffix(strings.TrimSpace(mount.Destination), "/")
		if containerPath == destination || strings.HasPrefix(containerPath, destination+"/") {
			return true
		}
	}
	return false
}
