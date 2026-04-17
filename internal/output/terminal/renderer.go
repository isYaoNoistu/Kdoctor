package terminal

import (
	"bytes"
	"fmt"
	"strings"

	"kdoctor/internal/localize"
	"kdoctor/pkg/model"
)

type Renderer struct{}

func (Renderer) Render(report model.Report) ([]byte, error) {
	var buf bytes.Buffer

	profile := strings.TrimSpace(report.Profile)
	if profile == "" {
		profile = "未指定"
	}

	fmt.Fprintf(&buf, "模式：%s\n", report.Mode)
	fmt.Fprintf(&buf, "配置模板：%s\n", profile)
	fmt.Fprintf(&buf, "总体状态：%s\n", localize.TranslateStatus(report.Summary.Status))
	fmt.Fprintf(&buf, "检查时间：%s\n", report.CheckedAt.Format("2006-01-02 15:04:05Z07:00"))
	fmt.Fprintf(&buf, "耗时：%dms\n", report.ElapsedMs)
	fmt.Fprintf(&buf, "Broker 存活：%d/%d\n", report.Summary.BrokerAlive, report.Summary.BrokerTotal)

	if report.Summary.Overview != "" {
		fmt.Fprintf(&buf, "概览：%s\n", report.Summary.Overview)
	}
	if len(report.Summary.RootCauses) > 0 {
		buf.WriteString("主因判断：\n")
		for _, cause := range report.Summary.RootCauses {
			fmt.Fprintf(&buf, "- %s\n", cause)
		}
	}
	if len(report.Summary.RecommendedActions) > 0 {
		buf.WriteString("建议动作：\n")
		for _, action := range report.Summary.RecommendedActions {
			fmt.Fprintf(&buf, "- %s\n", action)
		}
	}

	buf.WriteString("\n检查结果：\n")
	for _, check := range report.Checks {
		fmt.Fprintf(&buf, "- [%s] %s %s：%s\n", localize.TranslateStatus(check.Status), check.ID, check.Module, check.Summary)
		if len(check.Evidence) > 0 {
			fmt.Fprintf(&buf, "  证据：%s\n", strings.Join(check.Evidence, " | "))
		}
		if len(check.PossibleCauses) > 0 {
			fmt.Fprintf(&buf, "  可能原因：%s\n", strings.Join(check.PossibleCauses, " | "))
		}
		if len(check.NextActions) > 0 {
			fmt.Fprintf(&buf, "  下一步：%s\n", strings.Join(check.NextActions, " | "))
		}
	}

	if len(report.Errors) > 0 {
		buf.WriteString("\n附加错误：\n")
		for _, err := range report.Errors {
			fmt.Fprintf(&buf, "- %s\n", err)
		}
	}

	return buf.Bytes(), nil
}
