package producer

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type AcksChecker struct {
	Acks               string
	ExpectedDurability string
	MinISR             int
}

func (AcksChecker) ID() string     { return "PRD-001" }
func (AcksChecker) Name() string   { return "acks_minisr_safety" }
func (AcksChecker) Module() string { return "producer" }

func (c AcksChecker) Run(_ context.Context, _ *snapshot.Bundle) model.CheckResult {
	acks := strings.ToLower(strings.TrimSpace(c.Acks))
	if acks == "" {
		return rule.NewSkip("PRD-001", "acks_minisr_safety", "producer", "当前 profile 未提供 producer acks，暂不评估写入一致性配置")
	}

	evidence := []string{fmt.Sprintf("acks=%s", acks)}
	if c.MinISR > 0 {
		evidence = append(evidence, fmt.Sprintf("expected_min_isr=%d", c.MinISR))
	}
	if strings.EqualFold(strings.TrimSpace(c.ExpectedDurability), "strong") && acks != "all" {
		result := rule.NewFail("PRD-001", "acks_minisr_safety", "producer", "业务声明需要强一致，但 producer acks 并未使用 all")
		result.Evidence = evidence
		result.NextActions = []string{"将 producer acks 调整为 all", "确认业务对强一致写入的真实要求", "结合 min.insync.replicas 一起校验写入安全边界"}
		return result
	}
	if acks == "all" && c.MinISR <= 1 {
		result := rule.NewWarn("PRD-001", "acks_minisr_safety", "producer", "producer 虽使用 acks=all，但 min.insync.replicas 余量偏低，强一致保障有限")
		result.Evidence = evidence
		result.NextActions = []string{"提升 min.insync.replicas 至更合理值", "确认 RF 与 ISR 策略能支撑 acks=all", "不要把 acks=all 误解为天然强一致"}
		return result
	}

	result := rule.NewPass("PRD-001", "acks_minisr_safety", "producer", "producer acks 与当前一致性预期基本匹配")
	result.Evidence = evidence
	return result
}
