package kraft

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type MajorityChecker struct{}

func (MajorityChecker) ID() string     { return "KRF-004" }
func (MajorityChecker) Name() string   { return "controller_quorum_majority_evidence" }
func (MajorityChecker) Module() string { return "kraft" }

func (MajorityChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Network == nil || len(bundle.Network.ControllerChecks) == 0 {
		return rule.NewSkip("KRF-004", "controller_quorum_majority_evidence", "kraft", "controller quorum endpoints are not available for majority evaluation")
	}

	reachable := 0
	evidence := make([]string, 0, len(bundle.Network.ControllerChecks)+1)
	for _, check := range bundle.Network.ControllerChecks {
		if check.Reachable {
			reachable++
			evidence = append(evidence, fmt.Sprintf("controller=%s reachable duration_ms=%d", check.Address, check.DurationMs))
			continue
		}
		evidence = append(evidence, fmt.Sprintf("controller=%s unreachable error=%s", check.Address, check.Error))
	}

	majority := len(bundle.Network.ControllerChecks)/2 + 1
	evidence = append(evidence, fmt.Sprintf("reachable=%d majority=%d", reachable, majority))
	if activeCount, ok, activeEvidence := metricsAggregateMax(bundle, "kafka_controller_kafkacontroller_activecontrollercount"); ok {
		evidence = append(evidence, activeEvidence...)
		evidence = append(evidence, fmt.Sprintf("active_controller_count=%.0f", activeCount))
	}

	result := rule.NewPass("KRF-004", "controller_quorum_majority_evidence", "kraft", "controller quorum majority evidence is healthy")
	result.Evidence = evidence
	switch {
	case reachable < majority:
		result = rule.NewCrit("KRF-004", "controller_quorum_majority_evidence", "kraft", "controller quorum does not currently have majority evidence")
		result.Evidence = evidence
		result.NextActions = []string{"verify controller listener reachability between quorum voters", "check controller processes and recent election errors", "confirm controller.quorum.voters still reflects the active topology"}
	case reachable < len(bundle.Network.ControllerChecks):
		result = rule.NewWarn("KRF-004", "controller_quorum_majority_evidence", "kraft", "controller quorum still has majority but part of the voter set is unreachable")
		result.Evidence = evidence
		result.NextActions = []string{"stabilize the unreachable controller listeners before another failure happens", "compare reachability from the current execution host and broker host", "inspect controller logs for intermittent network or append failures"}
	}
	return result
}
