package lint

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type TopologyChecker struct {
	ExpectedBrokerCount int
	ExpectedControllers int
}

func (TopologyChecker) ID() string     { return "CFG-012" }
func (TopologyChecker) Name() string   { return "profile_compose_topology" }
func (TopologyChecker) Module() string { return "lint" }

func (c TopologyChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	services := kafkaServices(getCompose(snap))
	if len(services) == 0 {
		return rule.NewSkip("CFG-012", "profile_compose_topology", "lint", "compose Kafka services not available")
	}

	controllerCount := 0
	for _, service := range services {
		if contains(service.ProcessRoles, "controller") {
			controllerCount++
		}
	}
	evidence := []string{
		fmt.Sprintf("compose_broker_count=%d", len(services)),
		fmt.Sprintf("compose_controller_count=%d", controllerCount),
	}
	if c.ExpectedBrokerCount > 0 {
		evidence = append(evidence, fmt.Sprintf("profile_broker_count=%d", c.ExpectedBrokerCount))
	}
	if c.ExpectedControllers > 0 {
		evidence = append(evidence, fmt.Sprintf("profile_controller_count=%d", c.ExpectedControllers))
	}

	result := rule.NewPass("CFG-012", "profile_compose_topology", "lint", "profile expectations and compose topology are structurally aligned")
	result.Evidence = evidence
	if (c.ExpectedBrokerCount > 0 && c.ExpectedBrokerCount != len(services)) || (c.ExpectedControllers > 0 && c.ExpectedControllers != controllerCount) {
		result = rule.NewWarn("CFG-012", "profile_compose_topology", "lint", "profile topology expectations do not fully match the compose topology")
		result.Evidence = evidence
		result.NextActions = []string{"update the profile broker or controller count to match the active deployment", "verify whether compose or profile drifted after an environment change", "keep profile, compose, and runtime topology aligned to reduce diagnostic noise"}
	}
	return result
}
