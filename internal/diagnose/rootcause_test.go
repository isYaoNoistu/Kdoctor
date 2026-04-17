package diagnose

import (
	"strings"
	"testing"
	"time"

	"kdoctor/pkg/model"
)

func TestRootCauseDiagnosePrefersMetadataEndpointRootCause(t *testing.T) {
	report := model.NewReport(model.ModeProbe, "generic-bootstrap", time.Now())
	report.Checks = []model.CheckResult{
		{
			ID:          "NET-001",
			Status:      model.StatusPass,
			Summary:     "bootstrap endpoint reachable",
			NextActions: []string{"verify bootstrap endpoints"},
		},
		{
			ID:          "NET-003",
			Status:      model.StatusFail,
			Summary:     "metadata returned unreachable broker endpoints",
			NextActions: []string{"verify advertised.listeners"},
		},
		{
			ID:          "CLI-005",
			Status:      model.StatusFail,
			Summary:     "end-to-end probe failed",
			NextActions: []string{"check the failing stage first"},
		},
	}
	report.Finalize()

	RootCause{MaxCauses: 3, EnableConfidence: true}.Diagnose(&report)

	if len(report.Summary.RootCauses) == 0 {
		t.Fatalf("expected root causes to be generated")
	}
	if !strings.Contains(report.Summary.RootCauses[0], "metadata") && !strings.Contains(report.Summary.RootCauses[0], "advertised.listeners") {
		t.Fatalf("expected metadata endpoint root cause first, got %q", report.Summary.RootCauses[0])
	}
	if !strings.Contains(report.Summary.RootCauses[0], "反证/局限") {
		t.Fatalf("expected counter-evidence to be included, got %q", report.Summary.RootCauses[0])
	}
}

func TestIncidentSummarizeOverridesOverviewInIncidentMode(t *testing.T) {
	report := model.NewReport(model.ModeIncident, "generic-bootstrap", time.Now())
	report.Summary.RootCauses = []string{"高置信度主因：metadata 返回的 broker 地址对当前客户端不可达。"}
	report.Summary.RecommendedActions = []string{"优先核对 advertised.listeners。"}

	Incident{}.Summarize(&report)

	if !strings.Contains(report.Summary.Overview, "已锁定主因") {
		t.Fatalf("expected incident overview to be condensed, got %q", report.Summary.Overview)
	}
}
