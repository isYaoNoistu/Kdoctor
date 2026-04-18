package producer

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type TimeoutChecker struct {
	DeliveryTimeoutMs int
	RequestTimeoutMs  int
	LingerMs          int
}

func (TimeoutChecker) ID() string     { return "PRD-003" }
func (TimeoutChecker) Name() string   { return "delivery_timeout_sanity" }
func (TimeoutChecker) Module() string { return "producer" }

func (c TimeoutChecker) Run(_ context.Context, _ *snapshot.Bundle) model.CheckResult {
	if c.DeliveryTimeoutMs == 0 && c.RequestTimeoutMs == 0 && c.LingerMs == 0 {
		return rule.NewSkip("PRD-003", "delivery_timeout_sanity", "producer", "当前 profile 未提供 producer 超时参数，暂不评估 delivery timeout 组合")
	}

	evidence := []string{
		fmt.Sprintf("delivery_timeout_ms=%d", c.DeliveryTimeoutMs),
		fmt.Sprintf("request_timeout_ms=%d", c.RequestTimeoutMs),
		fmt.Sprintf("linger_ms=%d", c.LingerMs),
	}
	if c.DeliveryTimeoutMs > 0 && c.RequestTimeoutMs > 0 && c.DeliveryTimeoutMs < c.RequestTimeoutMs+c.LingerMs {
		result := rule.NewFail("PRD-003", "delivery_timeout_sanity", "producer", "delivery.timeout.ms 小于 request.timeout.ms + linger.ms，属于明显不合理配置")
		result.Evidence = evidence
		result.NextActions = []string{"提升 delivery.timeout.ms", "避免 request timeout 与 linger 组合把交付超时挤压得过短", "让生产重试与 broker 处理时间保留足够余量"}
		return result
	}

	result := rule.NewPass("PRD-003", "delivery_timeout_sanity", "producer", "producer 超时参数组合基本合理")
	result.Evidence = evidence
	return result
}
