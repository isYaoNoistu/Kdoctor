package markdown

import (
	"strings"
	"testing"
	"time"

	"kdoctor/pkg/model"
)

func TestRendererRendersChineseMarkdownSections(t *testing.T) {
	report := model.NewReport(model.ModeProbe, "generic-bootstrap", time.Now())
	report.ElapsedMs = 123
	report.Summary.RootCauses = []string{"最可能主因：metadata 返回地址不可达。"}
	report.Summary.RecommendedActions = []string{"优先核对 advertised.listeners。"}
	report.Checks = []model.CheckResult{
		{
			ID:          "NET-003",
			Module:      "网络",
			Status:      model.StatusFail,
			Summary:     "metadata 返回了不可达的 broker 端点",
			Evidence:    []string{"192.168.1.1:9194 不可达"},
			NextActions: []string{"核对 advertised.listeners"},
		},
	}

	payload, err := Renderer{}.Render(report)
	if err != nil {
		t.Fatalf("render markdown: %v", err)
	}
	output := string(payload)
	if !strings.Contains(output, "## 主因判断") {
		t.Fatalf("expected root cause section, got %q", output)
	}
	if !strings.Contains(output, "## 检查项") {
		t.Fatalf("expected checks section, got %q", output)
	}
}
