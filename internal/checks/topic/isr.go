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

func (c ISRChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Topic == nil {
		return rule.NewError("TOP-005", "min_insync_replicas_conflict", "topic", "ISR health cannot be evaluated", "topic snapshot missing")
	}
	if c.MinISR <= 0 {
		c.MinISR = 1
	}

	failing := 0
	warn := 0
	evidence := []string{}
	for _, topic := range bundle.Topic.Topics {
		for _, partition := range topic.Partitions {
			switch {
			case len(partition.ISR) < c.MinISR:
				failing++
				evidence = append(evidence, fmt.Sprintf("主题=%s 分区=%d ISR=%d minISR=%d", topic.Name, partition.ID, len(partition.ISR), c.MinISR))
			case len(partition.ISR) < len(partition.Replicas):
				warn++
			}
		}
	}

	result := rule.NewPass("TOP-005", "min_insync_replicas_conflict", "topic", "ISR 满足 min.insync.replicas 要求")
	result.Evidence = evidence
	if failing > 0 {
		result = rule.NewFail("TOP-005", "min_insync_replicas_conflict", "topic", "部分分区低于 min.insync.replicas，acks=all 可能失败")
		result.Evidence = evidence
		result.NextActions = []string{"检查副本健康状态", "检查受影响 broker 与磁盘", "在 ISR 恢复前降低写入压力"}
	} else if warn > 0 {
		result = rule.NewWarn("TOP-005", "min_insync_replicas_conflict", "topic", "部分分区副本不足，但仍高于 min.insync.replicas")
		result.NextActions = []string{"持续观察 ISR 是否继续收缩", "排查 follower 复制延迟", "在业务高峰前恢复副本余量"}
	}
	return result
}
