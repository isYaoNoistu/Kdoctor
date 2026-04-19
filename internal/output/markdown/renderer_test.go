package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"kdoctor/pkg/model"
)

func TestRendererBootstrapOnlyScenario(t *testing.T) {
	report := sampleReport("quick", "generic-bootstrap", []model.CheckResult{
		{ID: "NET-003", Module: "网络", Status: model.StatusFail, Summary: "metadata 返回了不可达的 broker 端点", Evidence: []string{"192.168.1.1:9194 unreachable", "192.168.1.1:9196 unreachable"}, NextActions: []string{"核对 advertised.listeners"}},
		{ID: "KFK-001", Module: "Kafka", Status: model.StatusPass, Summary: "已成功获取集群 metadata"},
		{ID: "CFG-001", Module: "配置", Status: model.StatusSkip, Summary: "没有可用的 compose 快照"},
	})

	output, err := Renderer{MaxEvidenceItems: 8}.Render(report)
	if err != nil {
		t.Fatalf("render markdown: %v", err)
	}
	text := string(output)
	assertContainsAll(t, text,
		"# Kdoctor 检查报告",
		"## 摘要",
		"## 证据覆盖",
		"## 主因判断",
		"## 重点问题",
		"| 失败 | NET-003 | 网络 | metadata 返回了不可达的 broker 端点 |",
	)
	assertGolden(t, "bootstrap_only.md", text)
	if strings.Contains(text, "CFG-001 配置") {
		t.Fatalf("expected skip check to stay out of default markdown appendix, got %q", text)
	}
}

func TestRendererProbeOnlyScenario(t *testing.T) {
	report := sampleReport("probe", "generic-bootstrap", []model.CheckResult{
		{ID: "CLI-002", Module: "客户端", Status: model.StatusFail, Summary: "生产探针失败", Evidence: []string{"stage=produce", "error=leader unavailable", "topic=_kdoctor_probe"}, NextActions: []string{"优先检查生产失败阶段", "核对探针主题 leader"}},
		{ID: "CLI-005", Module: "客户端", Status: model.StatusSkip, Summary: "上游阶段失败，端到端探针未执行"},
	})

	output, err := Renderer{MaxEvidenceItems: 2}.Render(report)
	if err != nil {
		t.Fatalf("render markdown: %v", err)
	}
	text := string(output)
	assertContainsAll(t, text,
		"### CLI-002 客户端",
		"- 核心证据：",
		"其余 1 条已省略",
		"- 下一步：",
	)
	assertGolden(t, "probe_only.md", text)
}

func TestRendererComposeDockerLogsScenario(t *testing.T) {
	report := sampleReport("incident", "single-host-3broker-kraft-prod", []model.CheckResult{
		{ID: "DKR-003", Module: "Docker", Status: model.StatusFail, Summary: "部分 Kafka 容器发生过 OOMKilled", Evidence: []string{"container=kafka2 oom_killed=true"}, NextActions: []string{"检查容器内存限制与 JVM 堆大小"}},
		{ID: "LOG-001", Module: "日志", Status: model.StatusWarn, Summary: "日志来源已获取，但部分样本不足或不够新鲜，后续日志判断需要谨慎解释", Evidence: []string{"source=file:/data/kafka/server.log line_count=80 byte_count=4096 latest_timestamp=2026-04-19T20:58:00+08:00 freshness=2m0s sample_sufficient=false empty=false"}},
		{ID: "CFG-006", Module: "配置", Status: model.StatusPass, Summary: "listeners 与 advertised.listeners 结构一致"},
	})

	output, err := Renderer{MaxEvidenceItems: 8, Verbose: true}.Render(report)
	if err != nil {
		t.Fatalf("render markdown: %v", err)
	}
	text := string(output)
	assertContainsAll(t, text,
		"## 完整附录",
		"<details>",
		"### CFG-006 配置",
		"### DKR-003 Docker",
		"### LOG-001 日志",
	)
	assertGolden(t, "compose_logs_verbose.md", text)
}

func sampleReport(mode string, profile string, checks []model.CheckResult) model.Report {
	report := model.NewReport(mode, profile, time.Date(2026, 4, 19, 21, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	report.ElapsedMs = 1234
	report.Summary.BrokerAlive = 2
	report.Summary.BrokerTotal = 3
	report.Summary.Overview = "本次共执行 3 项检查，最高状态为 失败。已识别 1 个优先级较高的主因，请优先按建议动作顺序处理。"
	report.Summary.DataSourceCoverage = []string{
		"网络=已启用，已获取证据",
		"Compose=未纳入本次运行",
		"日志=已启用，未获取证据",
	}
	report.Summary.RootCauses = []string{"最可能主因：metadata 返回地址与当前客户端视角不匹配。"}
	report.Summary.RecommendedActions = []string{"优先核对 advertised.listeners 与当前客户端网络路径。"}
	report.Checks = checks
	report.Finalize()
	return report
}

func assertContainsAll(t *testing.T, text string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(text, fragment) {
			t.Fatalf("expected output to contain %q, got %q", fragment, text)
		}
	}
}

func assertGolden(t *testing.T, name string, got string) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name)
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}
	if got != string(want) {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got, string(want))
	}
}
