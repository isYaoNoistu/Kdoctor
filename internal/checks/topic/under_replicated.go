package topic

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type UnderReplicatedChecker struct {
	WarnCount int
}

func (UnderReplicatedChecker) ID() string     { return "TOP-006" }
func (UnderReplicatedChecker) Name() string   { return "under_replicated_partitions" }
func (UnderReplicatedChecker) Module() string { return "topic" }

func (c UnderReplicatedChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Topic == nil {
		return rule.NewSkip("TOP-006", "under_replicated_partitions", "topic", "当前没有可用的 topic 快照，无法评估 UnderReplicatedPartitions")
	}
	if c.WarnCount <= 0 {
		c.WarnCount = 1
	}

	count := 0
	evidence := []string{}
	for _, topic := range bundle.Topic.Topics {
		for _, partition := range topic.Partitions {
			if len(partition.ISR) < len(partition.Replicas) {
				count++
				evidence = append(evidence, fmt.Sprintf("%s partition %d ISR=%d replicas=%d", topic.Name, partition.ID, len(partition.ISR), len(partition.Replicas)))
			}
		}
	}

	result := rule.NewPass("TOP-006", "under_replicated_partitions", "topic", "topic 副本同步状态正常，未发现 UnderReplicatedPartitions")
	result.Evidence = evidence
	if count >= c.WarnCount {
		result = rule.NewWarn("TOP-006", "under_replicated_partitions", "topic", "检测到 UnderReplicatedPartitions，已经影响副本同步健康")
		result.Evidence = evidence
		result.NextActions = []string{"检查 follower broker 健康状态", "结合磁盘、网络和 controller 检查一起判断", "在写入压力升高前优先恢复副本同步"}
	}
	return result
}
