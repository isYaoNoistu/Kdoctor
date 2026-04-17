package network

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ListenerChecker struct{}

func (ListenerChecker) ID() string     { return "NET-002" }
func (ListenerChecker) Name() string   { return "configured_listener_probe" }
func (ListenerChecker) Module() string { return "network" }

func (ListenerChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Network == nil {
		return rule.NewError("NET-002", "configured_listener_probe", "network", "configured listeners cannot be evaluated", "network snapshot missing")
	}

	checks := append([]snapshot.EndpointCheck(nil), snap.Network.BootstrapChecks...)
	checks = append(checks, snap.Network.ControllerChecks...)
	if len(checks) == 0 {
		return rule.NewSkip("NET-002", "configured_listener_probe", "network", "no explicit listener endpoints were provided")
	}

	unreachable := 0
	privateControllerMisses := 0
	evidence := make([]string, 0, len(checks))
	for _, check := range checks {
		if check.Reachable {
			evidence = append(evidence, fmt.Sprintf("%s %s reachable in %dms", check.Kind, check.Address, check.DurationMs))
			continue
		}
		unreachable++
		evidence = append(evidence, fmt.Sprintf("%s %s unreachable: %s", check.Kind, check.Address, check.Error))
		if check.Kind == "controller" && isExternalProbeView(snap) && isPrivateEndpoint(check.Address) {
			privateControllerMisses++
		}
	}

	result := rule.NewPass("NET-002", "configured_listener_probe", "network", "explicit listener endpoints are reachable from the current execution view")
	result.Evidence = evidence
	if unreachable == 0 {
		return result
	}

	if unreachable == privateControllerMisses && privateControllerMisses > 0 {
		result = rule.NewWarn("NET-002", "configured_listener_probe", "network", "some internal controller listeners are not reachable from the current external execution view")
		result.Evidence = evidence
		result.NextActions = []string{"run kdoctor from the Kafka host or private network", "verify controller listeners from an internal vantage point", "use KRF-002 and KRF-003 as the current source of controller health"}
		return result
	}

	result = rule.NewFail("NET-002", "configured_listener_probe", "network", "some explicit listener endpoints are not reachable")
	result.Evidence = evidence
	result.NextActions = []string{"verify listener binding addresses", "verify firewall and port exposure", "compare the failing endpoints with compose and profile settings"}
	return result
}
