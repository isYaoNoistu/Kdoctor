package topic

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type UnderMinISRChecker struct {
	MinISR       int
	AtMinISRWarn int
}

func (UnderMinISRChecker) ID() string     { return "TOP-007" }
func (UnderMinISRChecker) Name() string   { return "under_min_isr" }
func (UnderMinISRChecker) Module() string { return "topic" }

func (c UnderMinISRChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Topic == nil {
		return rule.NewSkip("TOP-007", "under_min_isr", "topic", "当前没有可用的 topic 快照，无法评估 UnderMinISR/AtMinISR")
	}
	if c.MinISR <= 0 {
		c.MinISR = 1
	}
	if c.AtMinISRWarn <= 0 {
		c.AtMinISRWarn = 1
	}

	underMin := 0
	atMin := 0
	evidence := []string{}
	for _, topic := range bundle.Topic.Topics {
		for _, partition := range topic.Partitions {
			switch {
			case len(partition.ISR) < c.MinISR:
				underMin++
				evidence = append(evidence, fmt.Sprintf("主题=%s 分区=%d ISR=%d minISR=%d", topic.Name, partition.ID, len(partition.ISR), c.MinISR))
			case len(partition.ISR) == c.MinISR:
				atMin++
				evidence = append(evidence, fmt.Sprintf("主题=%s 分区=%d 已到 minISR=%d", topic.Name, partition.ID, c.MinISR))
			}
		}
	}

	result := rule.NewPass("TOP-007", "under_min_isr", "topic", "topic ISR 仍有安全余量，未发现 UnderMinISR/AtMinISR 风险")
	result.Evidence = evidence
	if underMin > 0 {
		result = rule.NewFail("TOP-007", "under_min_isr", "topic", "检测到 UnderMinISR，acks=all 写入已经存在失败风险")
		result.Evidence = evidence
		result.NextActions = []string{"优先恢复 ISR 中的 follower 副本", "降低写入压力直到 ISR 恢复", "结合客户端生产探针与 ISR/副本状态一起确认影响范围"}
		return result
	}
	if atMin >= c.AtMinISRWarn {
		result = rule.NewWarn("TOP-007", "under_min_isr", "topic", "检测到 AtMinISR，当前写入链路已经接近 min.insync.replicas 边界")
		result.Evidence = evidence
		result.NextActions = []string{"持续观察 ISR 恢复情况", "在流量高峰前优先排查复制延迟", "确认 follower broker 没有磁盘或网络压力"}
	}
	return result
}
