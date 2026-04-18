package consumer

import (
	"context"
	"fmt"

	"kdoctor/internal/config"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type LagChecker struct {
	WarnLag int64
	CritLag int64
	Targets []config.GroupProbeTarget
}

func (LagChecker) ID() string     { return "CSM-001" }
func (LagChecker) Name() string   { return "consumer_group_lag" }
func (LagChecker) Module() string { return "consumer" }

func (c LagChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	groups := groupSnap(bundle)
	if groups == nil || !groups.Collected {
		return rule.NewSkip("CSM-001", "consumer_group_lag", "consumer", "当前未配置消费组 lag 检查目标")
	}
	if len(groups.Targets) == 0 {
		if len(groups.Errors) > 0 {
			return rule.NewError("CSM-001", "consumer_group_lag", "consumer", "消费组 lag 无法评估", groups.Errors[0])
		}
		return rule.NewSkip("CSM-001", "consumer_group_lag", "consumer", "当前没有可用的消费组 lag 目标")
	}

	targetMap := buildTargetMap(c.Targets)
	result := rule.NewPass("CSM-001", "consumer_group_lag", "consumer", "目标消费组 lag 处于可接受范围")
	lagWarn := 0
	lagCrit := 0
	for _, target := range groups.Targets {
		warnLag, critLag := thresholdsForTarget(target, targetMap, c.WarnLag, c.CritLag)
		result.Evidence = append(result.Evidence, fmt.Sprintf("group_id=%s topic=%s state=%s lag=%d members=%d coordinator=%s missing_offsets=%d", target.GroupID, target.Topic, target.State, target.TotalLag, target.MemberCount, target.Coordinator, target.MissingOffsets))
		if target.Error != "" {
			result.Evidence = append(result.Evidence, fmt.Sprintf("group_id=%s error=%s", target.GroupID, target.Error))
			lagWarn++
			continue
		}
		if critLag > 0 && target.TotalLag >= critLag {
			lagCrit++
			continue
		}
		if warnLag > 0 && target.TotalLag >= warnLag {
			lagWarn++
		}
	}

	switch {
	case lagCrit > 0:
		result = rule.NewCrit("CSM-001", "consumer_group_lag", "consumer", "部分关键消费组 lag 已超过高危阈值")
		result.Evidence = append(result.Evidence, summarizeTargets(groups)...)
		result.NextActions = []string{"优先检查消费组处理能力和业务侧消费速度", "核对受影响 topic 的 leader、ISR 与 broker 健康", "确认是否存在 fetch throttle 或 rebalance 风暴"}
	case lagWarn > 0:
		result = rule.NewWarn("CSM-001", "consumer_group_lag", "consumer", "部分消费组 lag 偏高或位点视图不完整")
		result.Evidence = append(result.Evidence, summarizeTargets(groups)...)
		result.NextActions = []string{"持续观察 lag 变化趋势", "检查消费者处理耗时与并发度", "结合消费组状态与 coordinator 结果一起判断"}
	default:
		result.Evidence = append(result.Evidence, summarizeTargets(groups)...)
	}
	return result
}
