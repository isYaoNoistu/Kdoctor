package network

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type MetadataChecker struct{}

func (MetadataChecker) ID() string     { return "NET-003" }
func (MetadataChecker) Name() string   { return "metadata_endpoint_reachability" }
func (MetadataChecker) Module() string { return "network" }

func (MetadataChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Network == nil || len(snap.Network.MetadataChecks) == 0 {
		return rule.NewError("NET-003", "metadata_endpoint_reachability", "network", "metadata endpoints were not checked", "metadata endpoint checks missing")
	}

	unreachable := 0
	evidence := make([]string, 0, len(snap.Network.MetadataChecks))
	for _, check := range snap.Network.MetadataChecks {
		if !check.Reachable {
			unreachable++
			evidence = append(evidence, fmt.Sprintf("%s unreachable: %s", check.Address, check.Error))
			continue
		}
		evidence = append(evidence, fmt.Sprintf("%s reachable in %dms", check.Address, check.DurationMs))
	}

	result := rule.NewPass("NET-003", "metadata_endpoint_reachability", "network", "metadata returned broker endpoints are reachable")
	result.Evidence = evidence
	if unreachable > 0 {
		result = rule.NewFail("NET-003", "metadata_endpoint_reachability", "network", "metadata returned unreachable broker endpoints")
		result.Evidence = evidence
		result.NextActions = []string{"verify advertised.listeners", "verify returned broker ports are exposed", "verify routing from current client network"}
	}
	return result
}
