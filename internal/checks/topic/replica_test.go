package topic

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestReplicaHealthCheckerWarnsOnUnderReplicatedPartition(t *testing.T) {
	leader := int32(1)
	checker := ReplicaHealthChecker{}
	result := checker.Run(context.Background(), &snapshot.Bundle{
		Topic: &snapshot.TopicSnapshot{
			Topics: []snapshot.TopicInfo{
				{
					Name: "orders",
					Partitions: []snapshot.PartitionInfo{
						{ID: 0, LeaderID: &leader, Replicas: []int32{1, 2, 3}, ISR: []int32{1, 2}},
					},
				},
			},
		},
	})

	if result.Status != model.StatusWarn {
		t.Fatalf("expected WARN, got %s", result.Status)
	}
}
