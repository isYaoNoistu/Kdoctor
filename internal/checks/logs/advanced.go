package logs

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type HitContextChecker struct{}

func (HitContextChecker) ID() string     { return "LOG-005" }
func (HitContextChecker) Name() string   { return "known_error_hit_context" }
func (HitContextChecker) Module() string { return "logs" }

func (HitContextChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	logs := logSnap(bundle)
	if logs == nil || !logs.Collected || len(logs.Sources) == 0 {
		return rule.NewSkip("LOG-005", "known_error_hit_context", "logs", "缺少日志来源，无法评估关键错误指纹命中上下文")
	}
	if len(logs.Matches) == 0 {
		result := rule.NewPass("LOG-005", "known_error_hit_context", "logs", "当前日志窗口没有命中关键错误指纹")
		appendSourceSummary(&result, logs)
		return result
	}

	result := resultForMatchSeverity("LOG-005", "known_error_hit_context", "当前日志窗口已命中关键错误指纹，并已给出来源与次数上下文", highestSeverity(logs.Matches))
	appendSourceSummary(&result, logs)
	for _, match := range logs.Matches {
		result.Evidence = append(result.Evidence, fmt.Sprintf("%s 次数=%d 来源=%v 含义=%s", match.ID, match.Count, match.AffectedSources, match.Meaning))
	}
	return result
}

type FreshnessChecker struct{}

func (FreshnessChecker) ID() string     { return "LOG-006" }
func (FreshnessChecker) Name() string   { return "log_source_freshness" }
func (FreshnessChecker) Module() string { return "logs" }

func (FreshnessChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	logs := logSnap(bundle)
	if logs == nil || !logs.Collected || len(logs.Sources) == 0 {
		return rule.NewSkip("LOG-006", "log_source_freshness", "logs", "缺少日志来源，无法评估日志新鲜度与样本充分性")
	}

	stale, sparse, empty := logSourceIssues(logs)
	result := rule.NewPass("LOG-006", "log_source_freshness", "logs", "日志来源足够新鲜，且样本量基本充足")
	appendSourceEvidence(&result, logs)
	if stale > 0 || sparse > 0 || empty > 0 {
		result = rule.NewWarn("LOG-006", "log_source_freshness", "logs", "部分日志来源过旧、过少或为空，日志结论可信度受限")
		appendSourceEvidence(&result, logs)
		result.NextActions = []string{"扩大日志时间窗口或 tail lines", "确认日志目录、容器日志和 lookback 设置覆盖到故障时间段", "避免只根据稀疏日志做强结论"}
	}
	return result
}

type StormChecker struct {
	RepeatThreshold int
}

func (StormChecker) ID() string     { return "LOG-007" }
func (StormChecker) Name() string   { return "repeated_fingerprint_storm" }
func (StormChecker) Module() string { return "logs" }

func (c StormChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	logs := logSnap(bundle)
	if logs == nil || !logs.Collected || len(logs.Sources) == 0 {
		return rule.NewSkip("LOG-007", "repeated_fingerprint_storm", "logs", "缺少日志来源，无法评估指纹风暴")
	}
	if c.RepeatThreshold <= 0 {
		c.RepeatThreshold = 5
	}
	if len(logs.Matches) == 0 {
		return rule.NewPass("LOG-007", "repeated_fingerprint_storm", "logs", "当前日志窗口未见重复错误指纹风暴")
	}

	storm := false
	evidence := []string{}
	for _, match := range logs.Matches {
		evidence = append(evidence, fmt.Sprintf("%s count=%d sources=%d", match.ID, match.Count, len(match.AffectedSources)))
		if match.Count >= c.RepeatThreshold || len(match.AffectedSources) > 1 {
			storm = true
		}
	}

	result := rule.NewPass("LOG-007", "repeated_fingerprint_storm", "logs", "日志中虽有错误指纹，但未形成明显的重复风暴")
	result.Evidence = evidence
	if storm {
		result = rule.NewWarn("LOG-007", "repeated_fingerprint_storm", "logs", "日志中同类错误已经形成重复风暴，当前问题更像持续性故障而非一次性抖动")
		result.Evidence = evidence
		result.NextActions = []string{"优先处理反复出现的最高级别日志指纹", "把日志指纹与网络、controller、ISR 结果做交叉验证", "避免把持续性故障误判成瞬时波动"}
	}
	return result
}

type CustomPatternChecker struct{}

func (CustomPatternChecker) ID() string     { return "LOG-008" }
func (CustomPatternChecker) Name() string   { return "custom_pattern_library" }
func (CustomPatternChecker) Module() string { return "logs" }

func (CustomPatternChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	logs := logSnap(bundle)
	if logs == nil || !logs.Collected {
		return rule.NewSkip("LOG-008", "custom_pattern_library", "logs", "当前输入模式未启用日志采集")
	}
	if logs.CustomPatternCount > 0 {
		result := rule.NewPass("LOG-008", "custom_pattern_library", "logs", "已加载自定义日志规则库，可承载内部经验指纹")
		result.Evidence = []string{fmt.Sprintf("custom_pattern_count=%d", logs.CustomPatternCount)}
		return result
	}
	if hasCustomPatternWarning(logs.Warnings) {
		result := rule.NewWarn("LOG-008", "custom_pattern_library", "logs", "已配置自定义日志规则目录，但当前没有成功加载任何规则")
		result.Evidence = append([]string{}, logs.Warnings...)
		result.NextActions = []string{"检查 custom_patterns_dir 路径与文件格式", "确认规则文件使用 json/yaml 且能被正确解析", "把内部高频故障继续沉淀到规则库"}
		return result
	}
	return rule.NewSkip("LOG-008", "custom_pattern_library", "logs", "当前未配置或未使用自定义日志规则库")
}

func hasCustomPatternWarning(warnings []string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, "自定义日志指纹") {
			return true
		}
	}
	return false
}
