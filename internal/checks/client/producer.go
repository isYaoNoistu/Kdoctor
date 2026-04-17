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
	if result, skipped := skipIfProbeStageNotExecuted("CLI-002", "producer_probe", pickProbe(snap), snapshot.ProbeStageProduce); skipped {
		return result
	}

	result := rule.NewPass("CLI-002", "producer_probe", "client", "producer probe succeeded")
	result.Evidence = []string{
		fmt.Sprintf("topic=%s", snap.Probe.Topic),
		fmt.Sprintf("partition=%d offset=%d", snap.Probe.ProducedPartition, snap.Probe.ProducedOffset),
		fmt.Sprintf("produce_count=%d", snap.Probe.ProducedMessageCount),
		fmt.Sprintf("topic_created=%t", snap.Probe.TopicCreated),
		fmt.Sprintf("topic_ready_reason=%s", snap.Probe.TopicReadyReason),
		fmt.Sprintf("produce_duration_ms=%d", snap.Probe.ProduceDurationMs),
	}
	if !snap.Probe.ProduceOK {
		result = rule.NewFail("CLI-002", "producer_probe", "client", "producer probe failed")
		result.Evidence = mergeEvidence([]string{fmt.Sprintf("topic=%s", snap.Probe.Topic)}, probeEvidence(snap.Probe))
		result.NextActions = []string{"verify probe topic exists", "verify produce path and acks", "check ISR and leader health"}
	}
	return result
}
