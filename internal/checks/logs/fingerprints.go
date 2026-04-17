package logs

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type FingerprintChecker struct{}

func (FingerprintChecker) ID() string     { return "LOG-002" }
func (FingerprintChecker) Name() string   { return "error_fingerprints" }
func (FingerprintChecker) Module() string { return "logs" }

func (FingerprintChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	logs := logSnap(bundle)
	if logs == nil || !logs.Collected || len(logs.Sources) == 0 {
		return rule.NewSkip("LOG-002", "error_fingerprints", "logs", "缺少日志来源，无法评估日志指纹")
	}
	if len(logs.Matches) == 0 {
		result := rule.NewPass("LOG-002", "error_fingerprints", "logs", "近期日志未命中内置已知 Kafka 错误指纹")
		result.Evidence = append(result.Evidence, fmt.Sprintf("内置指纹数=%d", logs.BuiltinPatternCount))
		if logs.CustomPatternCount > 0 {
			result.Evidence = append(result.Evidence, fmt.Sprintf("自定义指纹数=%d", logs.CustomPatternCount))
		}
		appendSourceSummary(&result, logs)
		return result
	}

	result := resultForMatchSeverity("LOG-002", "error_fingerprints", "近期日志命中了已知 Kafka 错误指纹", highestSeverity(logs.Matches))
	result.Evidence = append(result.Evidence, fmt.Sprintf("内置指纹数=%d", logs.BuiltinPatternCount))
	if logs.CustomPatternCount > 0 {
		result.Evidence = append(result.Evidence, fmt.Sprintf("自定义指纹数=%d", logs.CustomPatternCount))
	}
	appendSourceSummary(&result, logs)
	for _, match := range logs.Matches {
		library := "内置"
		if match.Library == "custom" {
			library = "自定义"
		}
		result.Evidence = append(result.Evidence, fmt.Sprintf("%s 规则库=%s 严重级别=%s 次数=%d 来源=%v", match.ID, library, match.Severity, match.Count, match.AffectedSources))
		if len(result.PossibleCauses) < 4 {
			result.PossibleCauses = append(result.PossibleCauses, match.ProbableCauses...)
		}
		if len(result.NextActions) < 4 {
			result.NextActions = append(result.NextActions, match.RecommendedChecks...)
		}
	}
	return result
}
