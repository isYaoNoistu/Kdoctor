package consumer

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type OffsetResetChecker struct {
	AutoOffsetReset string
}

func (OffsetResetChecker) ID() string     { return "CSM-005" }
func (OffsetResetChecker) Name() string   { return "auto_offset_reset_semantics" }
func (OffsetResetChecker) Module() string { return "consumer" }

func (c OffsetResetChecker) Run(_ context.Context, _ *snapshot.Bundle) model.CheckResult {
	mode := strings.ToLower(strings.TrimSpace(c.AutoOffsetReset))
	if mode == "" {
		return rule.NewSkip("CSM-005", "auto_offset_reset_semantics", "consumer", "当前 profile 未提供 auto.offset.reset，暂不评估位点回退语义")
	}

	result := rule.NewPass("CSM-005", "auto_offset_reset_semantics", "consumer", "auto.offset.reset 已显式声明")
	result.Evidence = []string{fmt.Sprintf("auto_offset_reset=%s", mode)}
	switch mode {
	case "latest":
		result = rule.NewWarn("CSM-005", "auto_offset_reset_semantics", "consumer", "auto.offset.reset=latest，新消费组不会读取历史积压消息")
		result.Evidence = []string{"auto_offset_reset=latest"}
		result.NextActions = []string{"确认这是否符合业务预期", "如需要回补历史数据，请评估 earliest 或显式位点初始化", "避免把“消费不到旧消息”误判成 Kafka 故障"}
	case "none":
		result = rule.NewWarn("CSM-005", "auto_offset_reset_semantics", "consumer", "auto.offset.reset=none，缺失位点时消费者会直接报错")
		result.Evidence = []string{"auto_offset_reset=none"}
		result.NextActions = []string{"确认业务能处理缺失位点异常", "如需自动回退，请改为 earliest/latest", "结合消费组创建流程一起评估"}
	}
	return result
}
