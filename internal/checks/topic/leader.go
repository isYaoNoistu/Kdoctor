package topic

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type LeaderChecker struct{}

func (LeaderChecker) ID() string     { return "TOP-003" }
func (LeaderChecker) Name() string   { return "partition_leader_presence" }
func (LeaderChecker) Module() string { return "topic" }

func (LeaderChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Topic == nil {
		return rule.NewError("TOP-003", "partition_leader_presence", "topic", "topic leadership cannot be evaluated", "topic snapshot missing")
	}

	missing := 0
	evidence := []string{}
	for _, topic := range snap.Topic.Topics {
		for _, partition := range topic.Partitions {
			if partition.LeaderID == nil {
				missing++
				evidence = append(evidence, fmt.Sprintf("%s partition %d has no leader", topic.Name, partition.ID))
			}
		}
	}
	result := rule.NewPass("TOP-003", "partition_leader_presence", "topic", "all partitions have leaders")
	result.Evidence = evidence
	if missing > 0 {
		result = rule.NewFail("TOP-003", "partition_leader_presence", "topic", "partitions without leader detected")
		result.Evidence = evidence
		result.NextActions = []string{"verify controller health", "verify affected brokers are online", "check partition reassignment or recent broker failures"}
	}
	return result
}
