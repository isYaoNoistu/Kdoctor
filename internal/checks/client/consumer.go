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
	if snap == nil || snap.Probe == nil {
		return rule.NewError("CLI-003", "consumer_probe", "client", "probe snapshot missing", "probe snapshot missing")
	}
	if snap.Probe.Skipped {
		return rule.NewSkip("CLI-003", "consumer_probe", "client", snap.Probe.Reason)
	}

	result := rule.NewPass("CLI-003", "consumer_probe", "client", "consumer probe succeeded")
	result.Evidence = []string{
		fmt.Sprintf("message_id=%s", snap.Probe.MessageID),
		fmt.Sprintf("consume_duration_ms=%d", snap.Probe.ConsumeDurationMs),
	}
	if !snap.Probe.ConsumeOK {
		result = rule.NewFail("CLI-003", "consumer_probe", "client", "consumer probe failed")
		result.Evidence = []string{
			fmt.Sprintf("message_id=%s", snap.Probe.MessageID),
			fmt.Sprintf("failure_stage=%s", snap.Probe.FailureStage),
			fmt.Sprintf("error=%s", snap.Probe.Error),
		}
		result.NextActions = []string{"verify topic leader and offsets", "verify fetch path from current client network", "check consumer side timeouts"}
	}
	return result
}
