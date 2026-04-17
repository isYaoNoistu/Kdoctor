package client

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ConsumerChecker struct{}

func (ConsumerChecker) ID() string     { return "CLI-003" }
func (ConsumerChecker) Name() string   { return "consumer_probe" }
func (ConsumerChecker) Module() string { return "client" }

func (ConsumerChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if result, skipped := skipIfProbeStageNotExecuted("CLI-003", "consumer_probe", pickProbe(snap), snapshot.ProbeStageConsume); skipped {
		return result
	}

	result := rule.NewPass("CLI-003", "consumer_probe", "client", "consumer probe succeeded")
	result.Evidence = []string{
		fmt.Sprintf("message_id=%s", snap.Probe.MessageID),
		fmt.Sprintf("executed_stage=%s", snap.Probe.ExecutedStage),
		fmt.Sprintf("consume_duration_ms=%d", snap.Probe.ConsumeDurationMs),
	}
	if !snap.Probe.ConsumeOK {
		result = rule.NewFail("CLI-003", "consumer_probe", "client", "consumer probe failed")
		result.Evidence = mergeEvidence([]string{fmt.Sprintf("message_id=%s", snap.Probe.MessageID)}, probeEvidence(snap.Probe))
		result.NextActions = []string{"verify topic leader and offsets", "verify fetch path from current client network", "check consumer side timeouts"}
	}
	return result
}
