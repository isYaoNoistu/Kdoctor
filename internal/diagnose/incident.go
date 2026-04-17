package diagnose

import (
	"fmt"
	"strings"

	"kdoctor/pkg/model"
)

type Incident struct{}

func (Incident) Summarize(report *model.Report) {
	if report == nil || report.Mode != model.ModeIncident {
		return
	}

	sections := []string{}
	if len(report.Summary.RootCauses) > 0 {
		sections = append(sections, fmt.Sprintf("已锁定主因：%s", report.Summary.RootCauses[0]))
	}
	if len(report.Summary.RecommendedActions) > 0 {
		sections = append(sections, fmt.Sprintf("建议优先动作：%s", report.Summary.RecommendedActions[0]))
	}
	if len(sections) == 0 {
		sections = append(sections, "当前没有足够的故障证据形成 incident 摘要，请结合检查明细继续排查。")
	}

	report.Summary.Overview = strings.Join(sections, "；")
}
