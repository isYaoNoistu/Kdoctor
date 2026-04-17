package docker

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ExistenceChecker struct{}

func (ExistenceChecker) ID() string     { return "DKR-001" }
func (ExistenceChecker) Name() string   { return "container_existence" }
func (ExistenceChecker) Module() string { return "docker" }

func (ExistenceChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	docker := dockerSnap(bundle)
	if docker == nil || !docker.Collected {
		return rule.NewSkip("DKR-001", "container_existence", "docker", "docker runtime is not enabled in the current input mode")
	}
	if !docker.Available {
		result := rule.NewSkip("DKR-001", "container_existence", "docker", "docker runtime is not available on the current execution host")
		result.Evidence = append(result.Evidence, docker.Errors...)
		return result
	}

	missing := 0
	containers := dockerContainerMap(docker)
	evidence := []string{}
	for _, name := range docker.ExpectedNames {
		container, ok := containers[name]
		if !ok || container.State == "" && container.Status == "" && container.Image == "" {
			missing++
			evidence = append(evidence, fmt.Sprintf("%s missing", name))
			continue
		}
		evidence = append(evidence, fmt.Sprintf("%s present", name))
	}

	result := rule.NewPass("DKR-001", "container_existence", "docker", "all expected Kafka containers exist")
	result.Evidence = evidence
	if missing > 0 {
		result = rule.NewFail("DKR-001", "container_existence", "docker", "some expected Kafka containers do not exist")
		result.Evidence = evidence
		result.NextActions = []string{"verify compose service names and container_names", "start the missing containers", "check whether the execution host is the intended Docker host"}
	}
	return result
}
