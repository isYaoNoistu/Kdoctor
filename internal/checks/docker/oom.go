package docker

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type OOMChecker struct{}

func (OOMChecker) ID() string     { return "DKR-003" }
func (OOMChecker) Name() string   { return "container_oomkilled" }
func (OOMChecker) Module() string { return "docker" }

func (OOMChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	docker := dockerSnap(bundle)
	if docker == nil || !docker.Collected {
		return rule.NewSkip("DKR-003", "container_oomkilled", "docker", "docker runtime is not enabled in the current input mode")
	}
	if !docker.Available {
		result := rule.NewSkip("DKR-003", "container_oomkilled", "docker", "docker runtime is not available on the current execution host")
		result.Evidence = append(result.Evidence, docker.Errors...)
		return result
	}

	oomKilled := 0
	evidence := []string{}
	for _, container := range docker.Containers {
		if container.OOMKilled {
			oomKilled++
			evidence = append(evidence, fmt.Sprintf("%s OOMKilled=true restart_count=%d", container.Name, container.RestartCount))
		}
	}

	result := rule.NewPass("DKR-003", "container_oomkilled", "docker", "no Kafka container shows an OOMKilled runtime state")
	result.Evidence = evidence
	if oomKilled > 0 {
		result = rule.NewFail("DKR-003", "container_oomkilled", "docker", "some Kafka containers were OOMKilled")
		result.Evidence = evidence
		result.NextActions = []string{"review container memory limits and heap size", "inspect broker logs for memory pressure", "reduce load or restart carefully after adding headroom"}
	}
	return result
}
