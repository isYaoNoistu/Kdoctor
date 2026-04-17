package topic

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ReplicaHealthChecker struct{}

func (ReplicaHealthChecker) ID() string     { return "TOP-004" }
func (ReplicaHealthChecker) Name() string   { return "isr_health" }
func (ReplicaHealthChecker) Module() string { return "topic" }

func (ReplicaHealthChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Topic == nil {
		return rule.NewError("TOP-004", "isr_health", "topic", "ISR replica health cannot be evaluated", "topic snapshot missing")
	}

	underReplicated := 0
	emptyISR := 0
	evidence := []string{}
	for _, topic := range snap.Topic.Topics {
		for _, partition := range topic.Partitions {
			if len(partition.ISR) == 0 {
				emptyISR++
				evidence = append(evidence, fmt.Sprintf("%s partition %d has empty ISR", topic.Name, partition.ID))
				continue
			}
			if len(partition.ISR) < len(partition.Replicas) {
				underReplicated++
				evidence = append(evidence, fmt.Sprintf("%s partition %d ISR=%d replicas=%d", topic.Name, partition.ID, len(partition.ISR), len(partition.Replicas)))
			}
		}
	}

	result := rule.NewPass("TOP-004", "isr_health", "topic", "all partitions have full ISR")
	result.Evidence = evidence
	if emptyISR > 0 {
		result = rule.NewFail("TOP-004", "isr_health", "topic", "some partitions have empty ISR")
		result.Evidence = evidence
		result.NextActions = []string{"verify broker replication pipeline", "check affected broker health and disks", "check controller and partition movement"}
	} else if underReplicated > 0 {
		result = rule.NewWarn("TOP-004", "isr_health", "topic", "some partitions are under replicated")
		result.Evidence = evidence
		result.NextActions = []string{"verify follower brokers are healthy", "monitor ISR recovery before write pressure increases"}
	}
	return result
}
