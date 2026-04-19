package topic

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type OfflineReplicaChecker struct{}

func (OfflineReplicaChecker) ID() string     { return "TOP-008" }
func (OfflineReplicaChecker) Name() string   { return "offline_replicas" }
func (OfflineReplicaChecker) Module() string { return "topic" }

func (OfflineReplicaChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Topic == nil {
		return rule.NewSkip("TOP-008", "offline_replicas", "topic", "当前没有可用的 topic 快照，无法评估离线副本或离线分区")
	}

	leaderless := 0
	evidence := []string{}
	for _, topic := range bundle.Topic.Topics {
		for _, partition := range topic.Partitions {
			if partition.LeaderID == nil {
				leaderless++
				evidence = append(evidence, fmt.Sprintf("%s partition %d leader=nil", topic.Name, partition.ID))
			}
		}
	}

	result := rule.NewPass("TOP-008", "offline_replicas", "topic", "未发现离线副本或无 leader 分区")
	result.Evidence = evidence
	if leaderless > 0 {
		result = rule.NewFail("TOP-008", "offline_replicas", "topic", "检测到离线副本或无 leader 分区，属于高危故障")
		result.Evidence = evidence
		result.NextActions = []string{"优先确认受影响 broker 是否在线", "检查磁盘目录、挂载与 controller 状态", "查看 broker 日志中的 offline replica 或 leader election 错误"}
	}
	return result
}
