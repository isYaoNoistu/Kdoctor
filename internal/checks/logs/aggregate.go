package logs

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type AggregateChecker struct{}

func (AggregateChecker) ID() string     { return "LOG-004" }
func (AggregateChecker) Name() string   { return "duplicate_aggregation" }
func (AggregateChecker) Module() string { return "logs" }

func (AggregateChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	logs := logSnap(bundle)
	if logs == nil || !logs.Collected || len(logs.Sources) == 0 {
		return rule.NewSkip("LOG-004", "duplicate_aggregation", "logs", "缺少日志来源，无法做日志聚合")
	}
	if len(logs.Matches) == 0 {
		result := rule.NewPass("LOG-004", "duplicate_aggregation", "logs", "当前采集窗口内未观察到重复出现的已知错误指纹")
		appendSourceSummary(&result, logs)
		return result
	}

	result := rule.NewWarn("LOG-004", "duplicate_aggregation", "logs", "已汇总重复出现的日志指纹，可辅助定位持续性故障")
	appendSourceSummary(&result, logs)
	for _, match := range logs.Matches {
		result.Evidence = append(result.Evidence, fmt.Sprintf("%s 次数=%d 来源数=%d", match.ID, match.Count, len(match.AffectedSources)))
	}
	return result
}
