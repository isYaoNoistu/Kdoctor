package logs

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ExplanationChecker struct{}

func (ExplanationChecker) ID() string     { return "LOG-003" }
func (ExplanationChecker) Name() string   { return "error_explanations" }
func (ExplanationChecker) Module() string { return "logs" }

func (ExplanationChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	logs := logSnap(bundle)
	if logs == nil || !logs.Collected || len(logs.Sources) == 0 {
		return rule.NewSkip("LOG-003", "error_explanations", "logs", "缺少日志来源，无法生成日志解释")
	}
	if len(logs.Matches) == 0 {
		result := rule.NewPass("LOG-003", "error_explanations", "logs", "当前采集窗口内没有需要额外解释的日志指纹")
		appendSourceSummary(&result, logs)
		return result
	}

	result := rule.NewWarn("LOG-003", "error_explanations", "logs", "已生成日志分析摘要，并映射到可能原因与排查动作")
	appendSourceSummary(&result, logs)
	for _, match := range logs.Matches {
		result.Evidence = append(result.Evidence, fmt.Sprintf("%s 含义=%s", match.ID, match.Meaning))
		if len(result.PossibleCauses) < 5 {
			result.PossibleCauses = append(result.PossibleCauses, match.ProbableCauses...)
		}
		if len(result.NextActions) < 5 {
			result.NextActions = append(result.NextActions, match.RecommendedChecks...)
		}
	}
	return result
}
