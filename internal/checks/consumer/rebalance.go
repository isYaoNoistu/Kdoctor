package consumer

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type RebalanceChecker struct{}

func (RebalanceChecker) ID() string     { return "CSM-002" }
func (RebalanceChecker) Name() string   { return "consumer_group_rebalance" }
func (RebalanceChecker) Module() string { return "consumer" }

func (RebalanceChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	groups := groupSnap(bundle)
	if groups == nil || !groups.Collected {
		return rule.NewSkip("CSM-002", "consumer_group_rebalance", "consumer", "当前未配置消费组状态检查目标")
	}
	if len(groups.Targets) == 0 {
		return rule.NewSkip("CSM-002", "consumer_group_rebalance", "consumer", "当前没有可用的消费组状态目标")
	}

	rebalanceCount := 0
	deadCount := 0
	result := rule.NewPass("CSM-002", "consumer_group_rebalance", "consumer", "目标消费组状态整体稳定")
	for _, target := range groups.Targets {
		state := strings.TrimSpace(target.State)
		result.Evidence = append(result.Evidence, fmt.Sprintf("group_id=%s topic=%s state=%s members=%d", target.GroupID, target.Topic, state, target.MemberCount))
		switch {
		case strings.EqualFold(state, "Dead"):
			deadCount++
		case strings.Contains(strings.ToLower(state), "rebalance"):
			rebalanceCount++
		}
	}

	switch {
	case deadCount > 0:
		result = rule.NewFail("CSM-002", "consumer_group_rebalance", "consumer", "部分消费组已进入异常状态，可能影响持续消费")
		result.Evidence = append(result.Evidence, summarizeTargets(groups)...)
		result.NextActions = []string{"优先检查消费组成员是否持续在线", "检查消费者异常退出、session timeout 和 max.poll.interval 配置", "结合日志查看是否存在 rebalance 风暴"}
	case rebalanceCount > 0:
		result = rule.NewWarn("CSM-002", "consumer_group_rebalance", "consumer", "部分消费组处于 rebalance 状态，需要结合时间窗口持续观察")
		result.Evidence = append(result.Evidence, summarizeTargets(groups)...)
		result.NextActions = []string{"观察消费组状态是否能回到 Stable", "检查成员波动、心跳超时和业务处理耗时", "结合 lag 与日志进一步判断是否是持续性 rebalance"}
	default:
		result.Evidence = append(result.Evidence, summarizeTargets(groups)...)
	}
	return result
}
