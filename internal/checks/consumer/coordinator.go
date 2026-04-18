package consumer

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type CoordinatorChecker struct{}

func (CoordinatorChecker) ID() string     { return "CSM-006" }
func (CoordinatorChecker) Name() string   { return "consumer_group_coordinator" }
func (CoordinatorChecker) Module() string { return "consumer" }

func (CoordinatorChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	groups := groupSnap(bundle)
	if groups == nil || !groups.Collected {
		return rule.NewSkip("CSM-006", "consumer_group_coordinator", "consumer", "当前未配置消费组 coordinator 检查目标")
	}
	if len(groups.Targets) == 0 {
		if len(groups.Errors) > 0 {
			return rule.NewError("CSM-006", "consumer_group_coordinator", "consumer", "消费组 coordinator 无法评估", groups.Errors[0])
		}
		return rule.NewSkip("CSM-006", "consumer_group_coordinator", "consumer", "当前没有可用的消费组 coordinator 目标")
	}

	missingCoordinator := 0
	missingOffsets := 0
	result := rule.NewPass("CSM-006", "consumer_group_coordinator", "consumer", "消费组 coordinator 与位点视图可用")
	for _, target := range groups.Targets {
		result.Evidence = append(result.Evidence, fmt.Sprintf("group_id=%s topic=%s coordinator=%s missing_offsets=%d", target.GroupID, target.Topic, target.Coordinator, target.MissingOffsets))
		if target.Error != "" {
			result.Evidence = append(result.Evidence, fmt.Sprintf("group_id=%s error=%s", target.GroupID, target.Error))
		}
		if target.Coordinator == "" {
			missingCoordinator++
		}
		if target.MissingOffsets > 0 {
			missingOffsets++
		}
	}

	switch {
	case missingCoordinator > 0:
		result = rule.NewFail("CSM-006", "consumer_group_coordinator", "consumer", "部分消费组 coordinator 视图异常，位点与提交链路可能不稳定")
		result.Evidence = append(result.Evidence, summarizeTargets(groups)...)
		result.NextActions = []string{"检查 __consumer_offsets 与 coordinator 健康状态", "确认 group coordinator 是否可以正常选举与响应", "结合 commit probe 和 broker 日志继续定位"}
	case missingOffsets > 0:
		result = rule.NewWarn("CSM-006", "consumer_group_coordinator", "consumer", "部分消费组分区缺少已提交位点，需要结合 offset reset 语义判断")
		result.Evidence = append(result.Evidence, summarizeTargets(groups)...)
		result.NextActions = []string{"确认该消费组是否本就没有已提交位点", "检查 auto.offset.reset 与消费组初始化阶段", "结合业务预期判断是否是异常积压或新组"}
	default:
		result.Evidence = append(result.Evidence, summarizeTargets(groups)...)
	}
	return result
}
