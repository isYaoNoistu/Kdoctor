package topic

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ISRChecker struct {
	MinISR int
}

func (ISRChecker) ID() string     { return "TOP-005" }
func (ISRChecker) Name() string   { return "min_insync_replicas_conflict" }
func (ISRChecker) Module() string { return "topic" }

func (c ISRChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Topic == nil {
		return rule.NewError("TOP-005", "min_insync_replicas_conflict", "topic", "ISR health cannot be evaluated", "topic snapshot missing")
	}

	if c.MinISR <= 0 {
		c.MinISR = 1
	}

	risky := 0
	warnOnly := 0
	evidence := []string{}
	for _, topic := range snap.Topic.Topics {
		for _, partition := range topic.Partitions {
			if len(partition.ISR) < c.MinISR {
				risky++
				evidence = append(evidence, fmt.Sprintf("%s partition %d ISR=%d minISR=%d", topic.Name, partition.ID, len(partition.ISR), c.MinISR))
			} else if len(partition.ISR) < len(partition.Replicas) {
				warnOnly++
				evidence = append(evidence, fmt.Sprintf("%s partition %d under replicated ISR=%d replicas=%d", topic.Name, partition.ID, len(partition.ISR), len(partition.Replicas)))
			}
		}
	}

	result := rule.NewPass("TOP-005", "min_insync_replicas_conflict", "topic", "ISR satisfies min.insync.replicas")
	result.Evidence = evidence
	if risky > 0 {
		result = rule.NewFail("TOP-005", "min_insync_replicas_conflict", "topic", "some partitions are below min.insync.replicas and acks=all may fail")
		result.Evidence = evidence
		result.NextActions = []string{"verify replica health", "check affected brokers and disks", "reduce write pressure until ISR recovers"}
	} else if warnOnly > 0 {
		result = rule.NewWarn("TOP-005", "min_insync_replicas_conflict", "topic", "some partitions are under replicated but still above min.insync.replicas")
		result.Evidence = evidence
	}
	return result
}
