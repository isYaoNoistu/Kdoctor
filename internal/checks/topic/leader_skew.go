package topic

import (
	"context"
	"fmt"
	"math"
	"sort"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type LeaderSkewChecker struct {
	WarnPct float64
}

func (LeaderSkewChecker) ID() string     { return "TOP-009" }
func (LeaderSkewChecker) Name() string   { return "leader_skew" }
func (LeaderSkewChecker) Module() string { return "topic" }

func (c LeaderSkewChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Topic == nil {
		return rule.NewSkip("TOP-009", "leader_skew", "topic", "当前没有可用的 topic 快照，无法评估 leader 分布偏斜")
	}
	if c.WarnPct <= 0 {
		c.WarnPct = 30
	}

	leaderCount := map[int32]int{}
	totalLeaders := 0
	for _, topic := range bundle.Topic.Topics {
		for _, partition := range topic.Partitions {
			if partition.LeaderID == nil {
				continue
			}
			leaderCount[*partition.LeaderID]++
			totalLeaders++
		}
	}
	if totalLeaders == 0 || len(leaderCount) <= 1 {
		return rule.NewSkip("TOP-009", "leader_skew", "topic", "当前 leader 数据不足，暂不评估 leader 分布偏斜")
	}

	ids := make([]int, 0, len(leaderCount))
	maxLeaders := 0
	for id, count := range leaderCount {
		ids = append(ids, int(id))
		if count > maxLeaders {
			maxLeaders = count
		}
	}
	sort.Ints(ids)
	avg := float64(totalLeaders) / float64(len(leaderCount))
	threshold := avg * (1 + c.WarnPct/100)

	evidence := []string{fmt.Sprintf("total_leaders=%d avg_per_broker=%.2f threshold=%.2f", totalLeaders, avg, threshold)}
	for _, id := range ids {
		evidence = append(evidence, fmt.Sprintf("broker_id=%d leaders=%d", id, leaderCount[int32(id)]))
	}

	result := rule.NewPass("TOP-009", "leader_skew", "topic", "leader 分布相对均衡")
	result.Evidence = evidence
	if float64(maxLeaders) > math.Ceil(threshold) {
		result = rule.NewWarn("TOP-009", "leader_skew", "topic", "leader 分布明显偏向单个 broker，存在热点和抖动风险")
		result.Evidence = evidence
		result.NextActions = []string{"检查是否有长期 leader 倾斜或近期 broker 变更", "结合 broker 负载与请求延迟判断是否已形成热点", "必要时评估 leader rebalance 或分区重分配"}
	}
	return result
}
