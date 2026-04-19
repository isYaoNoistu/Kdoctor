package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"kdoctor/internal/localize"
	"kdoctor/pkg/model"
)

type Renderer struct {
	MaxEvidenceItems int
	ShowPassChecks   bool
	ShowSkipChecks   bool
	Verbose          bool
}

func (r Renderer) Render(report model.Report) ([]byte, error) {
	var buf bytes.Buffer

	profile := strings.TrimSpace(report.Profile)
	if profile == "" {
		profile = "未指定"
	}
	stats := summarizeChecks(report.Checks)
	problems := filterChecks(report.Checks, r.shouldShowProblem)
	appendix := filterChecks(report.Checks, r.shouldShowAppendix)

	buf.WriteString("# Kdoctor 检查报告\n\n")
	buf.WriteString("## 摘要\n\n")
	buf.WriteString("| 项目 | 值 |\n")
	buf.WriteString("| --- | --- |\n")
	fmt.Fprintf(&buf, "| 模式 | `%s` |\n", report.Mode)
	fmt.Fprintf(&buf, "| 配置模板 | `%s` |\n", profile)
	fmt.Fprintf(&buf, "| 总体状态 | `%s` |\n", localize.TranslateStatus(report.Summary.Status))
	fmt.Fprintf(&buf, "| 检查时间 | `%s` |\n", report.CheckedAt.Format("2006-01-02 15:04:05Z07:00"))
	fmt.Fprintf(&buf, "| 耗时 | `%dms` |\n", report.ElapsedMs)
	fmt.Fprintf(&buf, "| Broker 存活 | `%d/%d` |\n", report.Summary.BrokerAlive, report.Summary.BrokerTotal)
	fmt.Fprintf(&buf, "| 检查统计 | 严重 %d / 失败 %d / 告警 %d / 错误 %d / 通过 %d / 跳过 %d |\n\n",
		stats.crit, stats.fail, stats.warn, stats.err, stats.pass, stats.skip)

	if report.Summary.Overview != "" {
		fmt.Fprintf(&buf, "## 概览\n\n%s\n\n", report.Summary.Overview)
	}
	if len(report.Summary.DataSourceCoverage) > 0 {
		buf.WriteString("## 证据覆盖\n\n")
		buf.WriteString("| 来源 | 状态 |\n")
		buf.WriteString("| --- | --- |\n")
		for _, item := range report.Summary.DataSourceCoverage {
			name, state := splitCoverage(item)
			fmt.Fprintf(&buf, "| %s | %s |\n", name, state)
		}
		buf.WriteString("\n")
	}
	if len(report.Summary.DegradedTasks) > 0 {
		buf.WriteString("## 采集降级\n\n")
		for _, item := range report.Summary.DegradedTasks {
			fmt.Fprintf(&buf, "- %s\n", item)
		}
		buf.WriteString("\n")
	}
	if len(report.Summary.RootCauses) > 0 {
		buf.WriteString("## 主因判断\n\n")
		for _, cause := range report.Summary.RootCauses {
			fmt.Fprintf(&buf, "- %s\n", cause)
		}
		buf.WriteString("\n")
	}
	if len(report.Summary.RecommendedActions) > 0 {
		buf.WriteString("## 建议动作\n\n")
		for _, action := range report.Summary.RecommendedActions {
			fmt.Fprintf(&buf, "- %s\n", action)
		}
		buf.WriteString("\n")
	}

	buf.WriteString("## 重点问题\n\n")
	if len(problems) == 0 {
		buf.WriteString("当前没有需要立即处理的 FAIL/WARN/ERROR 明细。\n\n")
	} else {
		buf.WriteString("| 状态 | 编号 | 模块 | 摘要 |\n")
		buf.WriteString("| --- | --- | --- | --- |\n")
		for _, check := range problems {
			fmt.Fprintf(&buf, "| %s | %s | %s | %s |\n", localize.TranslateStatus(check.Status), check.ID, check.Module, check.Summary)
		}
		buf.WriteString("\n")

		buf.WriteString("## 重点问题详情\n\n")
		for _, check := range problems {
			renderCheck(&buf, check, r.MaxEvidenceItems)
		}
	}

	if len(appendix) > 0 {
		buf.WriteString("## 完整附录\n\n")
		buf.WriteString("<details>\n<summary>展开 PASS / SKIP 明细</summary>\n\n")
		for _, check := range appendix {
			renderCheck(&buf, check, r.MaxEvidenceItems)
		}
		buf.WriteString("</details>\n\n")
	}

	if len(report.Errors) > 0 {
		buf.WriteString("## 附加错误\n\n")
		for _, err := range report.Errors {
			fmt.Fprintf(&buf, "- %s\n", err)
		}
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

type checkStats struct {
	crit int
	fail int
	warn int
	err  int
	pass int
	skip int
}

func summarizeChecks(checks []model.CheckResult) checkStats {
	stats := checkStats{}
	for _, check := range checks {
		switch check.Status {
		case model.StatusCrit:
			stats.crit++
		case model.StatusFail:
			stats.fail++
		case model.StatusWarn:
			stats.warn++
		case model.StatusError, model.StatusTimeout:
			stats.err++
		case model.StatusSkip:
			stats.skip++
		case model.StatusPass:
			stats.pass++
		}
	}
	return stats
}

func splitCoverage(item string) (string, string) {
	parts := strings.SplitN(item, "=", 2)
	if len(parts) != 2 {
		return item, ""
	}
	return parts[0], parts[1]
}

func (r Renderer) shouldShowProblem(check model.CheckResult) bool {
	switch check.Status {
	case model.StatusCrit, model.StatusFail, model.StatusWarn, model.StatusError, model.StatusTimeout:
		return true
	default:
		return false
	}
}

func (r Renderer) shouldShowAppendix(check model.CheckResult) bool {
	if r.shouldShowProblem(check) {
		return false
	}
	if r.Verbose {
		return true
	}
	if r.ShowPassChecks && check.Status == model.StatusPass {
		return true
	}
	if r.ShowSkipChecks && check.Status == model.StatusSkip {
		return true
	}
	return false
}

func filterChecks(checks []model.CheckResult, predicate func(model.CheckResult) bool) []model.CheckResult {
	out := make([]model.CheckResult, 0, len(checks))
	for _, check := range checks {
		if predicate(check) {
			out = append(out, check)
		}
	}
	return out
}

func renderCheck(buf *bytes.Buffer, check model.CheckResult, maxEvidence int) {
	fmt.Fprintf(buf, "### %s %s\n\n", check.ID, check.Module)
	fmt.Fprintf(buf, "- 状态：`%s`\n", localize.TranslateStatus(check.Status))
	fmt.Fprintf(buf, "- 摘要：%s\n", check.Summary)
	renderList(buf, "核心证据", trimItems(check.Evidence, maxEvidence))
	risk := trimItems(check.PossibleCauses, maxEvidence)
	if check.Impact != "" {
		risk = append([]string{check.Impact}, risk...)
	}
	renderList(buf, "风险解释", trimItems(risk, maxEvidence))
	renderList(buf, "下一步", trimItems(check.NextActions, maxEvidence))
	buf.WriteString("\n")
}

func renderList(buf *bytes.Buffer, title string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(buf, "- %s：\n", title)
	for _, item := range items {
		fmt.Fprintf(buf, "  - %s\n", item)
	}
}

func trimItems(items []string, maxItems int) []string {
	if len(items) == 0 {
		return nil
	}
	if maxItems <= 0 || len(items) <= maxItems {
		return items
	}
	trimmed := append([]string(nil), items[:maxItems]...)
	trimmed = append(trimmed, fmt.Sprintf("其余 %d 条已省略", len(items)-maxItems))
	return trimmed
}
