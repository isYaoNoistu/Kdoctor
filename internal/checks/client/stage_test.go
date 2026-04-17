package client

import (
	"context"
	"strings"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestProducerCheckerSkipsWhenTopicReadyFailed(t *testing.T) {
	result := ProducerChecker{}.Run(context.Background(), &snapshot.Bundle{
		Probe: &snapshot.ProbeSnapshot{
			Topic:            "_kdoctor_probe",
			MetadataExecuted: true,
			MetadataOK:       true,
			FailureStage:     snapshot.ProbeStageTopicReady,
			ExecutedStage:    snapshot.ProbeStageTopicReady,
			TopicReadyReason: "probe topic could not be prepared",
			Error:            "ensure probe topic: create probe topic: not authorized",
		},
	})

	if result.Status != model.StatusSkip {
		t.Fatalf("expected SKIP, got %s", result.Status)
	}
	if !strings.Contains(result.Summary, "probe topic was not ready") {
		t.Fatalf("expected stage-aware skip summary, got %q", result.Summary)
	}
}

func TestConsumerCheckerSkipsWhenProduceFailed(t *testing.T) {
	result := ConsumerChecker{}.Run(context.Background(), &snapshot.Bundle{
		Probe: &snapshot.ProbeSnapshot{
			Topic:           "_kdoctor_probe",
			ProduceExecuted: true,
			FailureStage:    snapshot.ProbeStageProduce,
			ExecutedStage:   snapshot.ProbeStageProduce,
			Error:           "produce probe: send probe message: leader not available",
		},
	})

	if result.Status != model.StatusSkip {
		t.Fatalf("expected SKIP, got %s", result.Status)
	}
	if !strings.Contains(result.Summary, "produce stage failed") {
		t.Fatalf("expected produce-stage skip summary, got %q", result.Summary)
	}
}

func TestCommitCheckerSkipsWhenConsumeFailed(t *testing.T) {
	result := CommitChecker{}.Run(context.Background(), &snapshot.Bundle{
		Probe: &snapshot.ProbeSnapshot{
			Topic:           "_kdoctor_probe",
			ConsumeExecuted: true,
			FailureStage:    snapshot.ProbeStageConsume,
			ExecutedStage:   snapshot.ProbeStageConsume,
			Error:           "consume probe: timeout waiting for probe message",
		},
	})

	if result.Status != model.StatusSkip {
		t.Fatalf("expected SKIP, got %s", result.Status)
	}
	if !strings.Contains(result.Summary, "consume stage failed") {
		t.Fatalf("expected consume-stage skip summary, got %q", result.Summary)
	}
}

func TestEndToEndCheckerFailsWhenProbeStoppedAtProduce(t *testing.T) {
	result := EndToEndChecker{}.Run(context.Background(), &snapshot.Bundle{
		Probe: &snapshot.ProbeSnapshot{
			Topic:              "_kdoctor_probe",
			GroupID:            "group-1",
			MetadataOK:         true,
			FailureStage:       snapshot.ProbeStageProduce,
			ExecutedStage:      snapshot.ProbeStageProduce,
			Error:              "produce probe: send probe message: leader not available",
			EndToEndDurationMs: 1234,
		},
	})

	if result.Status != model.StatusFail {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}
	if len(result.Evidence) == 0 {
		t.Fatalf("expected evidence to include failure stage details")
	}
}
