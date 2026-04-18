package producer

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type IdempotenceChecker struct {
	EnableIdempotence bool
	Retries           int
	MaxInFlight       int
}

func (IdempotenceChecker) ID() string     { return "PRD-002" }
func (IdempotenceChecker) Name() string   { return "idempotence_retry_ordering" }
func (IdempotenceChecker) Module() string { return "producer" }

func (c IdempotenceChecker) Run(_ context.Context, _ *snapshot.Bundle) model.CheckResult {
	if c.Retries == 0 && c.MaxInFlight == 0 && !c.EnableIdempotence {
		return rule.NewSkip("PRD-002", "idempotence_retry_ordering", "producer", "当前 profile 未提供 producer 重试/幂等参数，暂不评估重复与乱序风险")
	}

	evidence := []string{
		fmt.Sprintf("enable_idempotence=%t", c.EnableIdempotence),
		fmt.Sprintf("retries=%d", c.Retries),
		fmt.Sprintf("max_in_flight=%d", c.MaxInFlight),
	}
	if !c.EnableIdempotence && c.Retries > 0 && c.MaxInFlight > 1 {
		result := rule.NewWarn("PRD-002", "idempotence_retry_ordering", "producer", "未启用幂等却同时开启重试和较大的 in-flight，存在重复或乱序风险")
		result.Evidence = evidence
		result.NextActions = []string{"优先启用 enable.idempotence", "如暂时不能启用幂等，至少收紧 max.in.flight", "确认业务是否能接受重复与乱序"}
		return result
	}

	result := rule.NewPass("PRD-002", "idempotence_retry_ordering", "producer", "当前 producer 幂等与重试组合未见明显风险")
	result.Evidence = evidence
	return result
}
