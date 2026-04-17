package docker

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type RunningChecker struct{}

func (RunningChecker) ID() string     { return "DKR-002" }
func (RunningChecker) Name() string   { return "container_running_state" }
func (RunningChecker) Module() string { return "docker" }

func (RunningChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	docker := dockerSnap(bundle)
	if docker == nil || !docker.Collected {
		return rule.NewSkip("DKR-002", "container_running_state", "docker", "docker runtime is not enabled in the current input mode")
	}
	if !docker.Available {
		result := rule.NewSkip("DKR-002", "container_running_state", "docker", "docker runtime is not available on the current execution host")
		result.Evidence = append(result.Evidence, docker.Errors...)
		return result
	}

	containers := dockerContainerMap(docker)
	notRunning := 0
	evidence := []string{}
	for _, name := range docker.ExpectedNames {
		container, ok := containers[name]
		if !ok {
			notRunning++
			evidence = append(evidence, fmt.Sprintf("%s not found", name))
			continue
		}
		if !container.Running {
			notRunning++
			evidence = append(evidence, fmt.Sprintf("%s state=%s status=%s", name, container.State, container.Status))
			continue
		}
		evidence = append(evidence, fmt.Sprintf("%s running status=%s", name, container.Status))
	}

	result := rule.NewPass("DKR-002", "container_running_state", "docker", "all expected Kafka containers are running")
	result.Evidence = evidence
	if notRunning > 0 {
		result = rule.NewFail("DKR-002", "container_running_state", "docker", "some expected Kafka containers are not running")
		result.Evidence = evidence
		result.NextActions = []string{"restart the stopped containers", "inspect docker logs for startup failures", "verify host resources and port conflicts"}
	}
	return result
}
