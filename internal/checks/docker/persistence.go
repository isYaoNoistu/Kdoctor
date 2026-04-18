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

type PersistenceChecker struct{}

func (PersistenceChecker) ID() string     { return "DKR-005" }
func (PersistenceChecker) Name() string   { return "runtime_mount_expectation" }
func (PersistenceChecker) Module() string { return "docker" }

func (PersistenceChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	docker := dockerSnap(bundle)
	if docker == nil || !docker.Collected {
		return rule.NewSkip("DKR-005", "runtime_mount_expectation", "docker", "docker runtime is not enabled in the current input mode")
	}
	if !docker.Available {
		result := rule.NewSkip("DKR-005", "runtime_mount_expectation", "docker", "docker runtime is not available on the current execution host")
		result.Evidence = append(result.Evidence, docker.Errors...)
		return result
	}

	services := composeutil.KafkaServices(getCompose(bundle))
	if len(services) == 0 {
		return rule.NewSkip("DKR-005", "runtime_mount_expectation", "docker", "compose Kafka services are not available for runtime mount validation")
	}

	containers := dockerContainerMap(docker)
	failures := 0
	warnings := 0
	evidence := []string{}
	for _, service := range services {
		containerName := service.ContainerName
		if strings.TrimSpace(containerName) == "" {
			containerName = service.ServiceName
		}

		container, ok := containers[containerName]
		if !ok {
			failures++
			evidence = append(evidence, fmt.Sprintf("service=%s container=%s missing", service.ServiceName, containerName))
			continue
		}

		requiredPaths := composeutil.ParseCSV(service.Environment["KAFKA_CFG_LOG_DIRS"])
		if metadata := strings.TrimSpace(service.Environment["KAFKA_CFG_METADATA_LOG_DIR"]); metadata != "" {
			requiredPaths = append(requiredPaths, metadata)
		}

		for _, requiredPath := range requiredPaths {
			mount, ok := coveringMount(requiredPath, container.Mounts)
			if !ok {
				failures++
				evidence = append(evidence, fmt.Sprintf("service=%s path=%s mounted=false", service.ServiceName, requiredPath))
				continue
			}

			evidence = append(evidence, fmt.Sprintf("service=%s path=%s source=%s destination=%s rw=%t", service.ServiceName, requiredPath, mount.Source, mount.Destination, mount.RW))
			if !mount.RW {
				warnings++
			}
		}
	}

	result := rule.NewPass("DKR-005", "runtime_mount_expectation", "docker", "docker inspect mounts match the expected Kafka storage paths")
	result.Evidence = evidence
	switch {
	case failures > 0:
		result = rule.NewFail("DKR-005", "runtime_mount_expectation", "docker", "some Kafka storage paths are not mounted in the current docker runtime view")
		result.Evidence = evidence
		result.NextActions = []string{"compare docker inspect mounts with compose volume declarations", "bind-mount Kafka data and metadata directories explicitly", "avoid leaving Kafka state only inside container layers"}
	case warnings > 0:
		result = rule.NewWarn("DKR-005", "runtime_mount_expectation", "docker", "Kafka storage mounts exist but part of the runtime mount set is read-only or unusual")
		result.Evidence = evidence
		result.NextActions = []string{"confirm Kafka data and metadata directories are mounted read-write", "check whether the current container mount policy matches the intended persistence design", "review recent container recreation or host-path changes"}
	}
	return result
}

func coveringMount(containerPath string, mounts []snapshot.DockerMount) (snapshot.DockerMount, bool) {
	containerPath = strings.TrimSuffix(strings.TrimSpace(containerPath), "/")
	best := snapshot.DockerMount{}
	bestLen := -1
	for _, mount := range mounts {
		destination := strings.TrimSuffix(strings.TrimSpace(mount.Destination), "/")
		if containerPath != destination && !strings.HasPrefix(containerPath, destination+"/") {
			continue
		}
		if len(destination) > bestLen {
			best = mount
			bestLen = len(destination)
		}
	}
	return best, bestLen >= 0
}
