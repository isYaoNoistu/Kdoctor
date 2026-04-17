package kafka

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestInternalTopicsCheckerWarnsWhenTransactionTopicMissing(t *testing.T) {
	leader := int32(1)
	checker := InternalTopicsChecker{}
	result := checker.Run(context.Background(), &snapshot.Bundle{
		Topic: &snapshot.TopicSnapshot{
			Topics: []snapshot.TopicInfo{
				{
					Name: "__consumer_offsets",
					Partitions: []snapshot.PartitionInfo{
						{ID: 0, LeaderID: &leader, Replicas: []int32{1, 2, 3}, ISR: []int32{1, 2, 3}},
					},
				},
			},
		},
	})

	if result.Status != model.StatusWarn {
		t.Fatalf("expected WARN, got %s", result.Status)
	}
}

func TestInternalTopicsCheckerWarnsWhenOffsetsTopicMissingBeforeCommit(t *testing.T) {
	checker := InternalTopicsChecker{}
	result := checker.Run(context.Background(), &snapshot.Bundle{
		Topic: &snapshot.TopicSnapshot{},
		Probe: &snapshot.ProbeSnapshot{
			FailureStage: snapshot.ProbeStageProduce,
		},
	})

	if result.Status != model.StatusWarn {
		t.Fatalf("expected WARN, got %s", result.Status)
	}
}

func TestInternalTopicsCheckerFailsWhenOffsetsTopicMissingAfterCommitExecution(t *testing.T) {
	checker := InternalTopicsChecker{}
	result := checker.Run(context.Background(), &snapshot.Bundle{
		Topic: &snapshot.TopicSnapshot{},
		Probe: &snapshot.ProbeSnapshot{
			CommitExecuted: true,
			CommitOK:       false,
			FailureStage:   snapshot.ProbeStageCommit,
		},
	})

	if result.Status != model.StatusFail {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}
}
