package client

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ProducerChecker struct{}

func (ProducerChecker) ID() string     { return "CLI-002" }
func (ProducerChecker) Name() string   { return "producer_probe" }
func (ProducerChecker) Module() string { return "client" }

func (ProducerChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Probe == nil {
		return rule.NewError("CLI-002", "producer_probe", "client", "probe snapshot missing", "probe snapshot missing")
	}
	if snap.Probe.Skipped {
		return rule.NewSkip("CLI-002", "producer_probe", "client", snap.Probe.Reason)
	}

	result := rule.NewPass("CLI-002", "producer_probe", "client", "producer probe succeeded")
	result.Evidence = []string{
		fmt.Sprintf("topic=%s", snap.Probe.Topic),
		fmt.Sprintf("partition=%d offset=%d", snap.Probe.ProducedPartition, snap.Probe.ProducedOffset),
		fmt.Sprintf("produce_duration_ms=%d", snap.Probe.ProduceDurationMs),
	}
	if !snap.Probe.ProduceOK {
		result = rule.NewFail("CLI-002", "producer_probe", "client", "producer probe failed")
		result.Evidence = []string{
			fmt.Sprintf("topic=%s", snap.Probe.Topic),
			fmt.Sprintf("failure_stage=%s", snap.Probe.FailureStage),
			fmt.Sprintf("error=%s", snap.Probe.Error),
		}
		result.NextActions = []string{"verify probe topic exists", "verify produce path and acks", "check ISR and leader health"}
	}
	return result
}
