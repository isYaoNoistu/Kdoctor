package logs

import (
	"context"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type SourcesChecker struct{}

func (SourcesChecker) ID() string     { return "LOG-001" }
func (SourcesChecker) Name() string   { return "log_sources" }
func (SourcesChecker) Module() string { return "logs" }

func (SourcesChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	logs := logSnap(bundle)
	if logs == nil || !logs.Collected {
		return rule.NewSkip("LOG-001", "log_sources", "logs", "当前输入模式未启用日志采集")
	}
	if len(logs.Sources) == 0 {
		result := rule.NewSkip("LOG-001", "log_sources", "logs", "当前执行视角没有可用的日志来源")
		result.Evidence = append(result.Evidence, logs.Errors...)
		return result
	}

	stale, sparse, empty := logSourceIssues(logs)
	switch {
	case empty > 0 || stale > 0 || sparse > 0 || len(logs.Warnings) > 0:
		result := rule.NewWarn("LOG-001", "log_sources", "logs", "日志来源已获取，但部分样本不足或不够新鲜，后续日志判断需要谨慎解释")
		appendSourceEvidence(&result, logs)
		return result
	default:
		result := rule.NewPass("LOG-001", "log_sources", "logs", "日志来源与样本质量满足本次分析需要")
		appendSourceEvidence(&result, logs)
		return result
	}
}
