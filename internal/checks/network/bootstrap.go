package network

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type BootstrapChecker struct{}

func (BootstrapChecker) ID() string     { return "NET-001" }
func (BootstrapChecker) Name() string   { return "bootstrap_tcp_connectivity" }
func (BootstrapChecker) Module() string { return "network" }

func (BootstrapChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Network == nil || len(snap.Network.BootstrapChecks) == 0 {
		return rule.NewError("NET-001", "bootstrap_tcp_connectivity", "network", "no bootstrap checks available", "network snapshot missing")
	}

	reachable := 0
	evidence := make([]string, 0, len(snap.Network.BootstrapChecks))
	for _, check := range snap.Network.BootstrapChecks {
		if check.Reachable {
			reachable++
		}
		if check.Error != "" {
			evidence = append(evidence, fmt.Sprintf("%s unreachable: %s", check.Address, check.Error))
			continue
		}
		evidence = append(evidence, fmt.Sprintf("%s reachable in %dms", check.Address, check.DurationMs))
	}

	result := rule.NewPass("NET-001", "bootstrap_tcp_connectivity", "network", "bootstrap endpoint reachable")
	result.Evidence = evidence
	if reachable == 0 {
		result = rule.NewFail("NET-001", "bootstrap_tcp_connectivity", "network", "no bootstrap endpoints reachable")
		result.Evidence = evidence
		result.NextActions = []string{"verify bootstrap addresses", "verify firewall and security group", "verify Kafka listener binding"}
	}
	return result
}
