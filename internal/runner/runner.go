package runner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kdoctor/internal/config"
	"kdoctor/internal/diagnose"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type Runner struct{}

func New() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, env *config.Runtime) (model.Report, error) {
	startedAt := time.Now()
	bundle, checks, errs := CollectAndCheck(ctx, env)
	degradedTasks, hardErrors := splitTaskMessages(errs)

	report := model.NewReport(env.Mode, env.ProfileName, startedAt)
	if bundle != nil && bundle.Kafka != nil {
		report.Summary.BrokerTotal = bundle.Kafka.ExpectedBrokerCount
		if report.Summary.BrokerTotal == 0 {
			report.Summary.BrokerTotal = len(bundle.Kafka.Brokers)
		}
		report.Summary.BrokerAlive = len(bundle.Kafka.Brokers)
		report.Summary.ControllerOK = bundle.Kafka.ControllerID != nil
	}
	report.Summary.DataSourceCoverage = summarizeCoverage(env, bundle)
	report.Summary.DegradedTasks = degradedTasks
	for _, check := range checks {
		report.AddCheck(check)
	}
	for _, err := range hardErrors {
		report.AddError(err)
	}
	report.ElapsedMs = time.Since(startedAt).Milliseconds()
	report.Finalize()
	diagnose.RootCause{
		MaxCauses:        env.DiagnosisMaxRootCauses,
		EnableConfidence: env.DiagnosisEnableConfidence,
	}.Diagnose(&report)
	diagnose.Incident{}.Summarize(&report)
	return report, nil
}

func splitTaskMessages(messages []string) ([]string, []string) {
	degraded := make([]string, 0)
	hardErrors := make([]string, 0)
	for _, msg := range messages {
		if isDegradedTaskMessage(msg) {
			degraded = append(degraded, msg)
			continue
		}
		hardErrors = append(hardErrors, msg)
	}
	return degraded, hardErrors
}

func isDegradedTaskMessage(message string) bool {
	return strings.Contains(message, "已降级")
}

func summarizeCoverage(env *config.Runtime, bundle *snapshot.Bundle) []string {
	coverage := make([]string, 0, 8)
	coverage = append(coverage, fmt.Sprintf("网络=%s", snapshotState(bundle != nil && bundle.Network != nil, "已采集", "缺失")))
	coverage = append(coverage, summarizeComposeCoverage(env, bundle))
	coverage = append(coverage, fmt.Sprintf("Kafka=%s", snapshotState(bundle != nil && bundle.Kafka != nil, "已采集", "缺失")))
	coverage = append(coverage, summarizeJMXCoverage(env, bundle))
	coverage = append(coverage, summarizeGroupCoverage(env, bundle))
	coverage = append(coverage, summarizeDockerCoverage(env, bundle))
	coverage = append(coverage, summarizeHostCoverage(env, bundle))
	coverage = append(coverage, summarizeLogCoverage(env, bundle))
	coverage = append(coverage, summarizeProbeCoverage(env, bundle))
	return coverage
}

func summarizeComposeCoverage(env *config.Runtime, bundle *snapshot.Bundle) string {
	if strings.TrimSpace(env.ComposePath) == "" {
		return "Compose=未提供"
	}
	return fmt.Sprintf("Compose=%s", snapshotState(bundle != nil && bundle.Compose != nil, "已采集", "缺失或解析失败"))
}

func summarizeGroupCoverage(env *config.Runtime, bundle *snapshot.Bundle) string {
	if len(env.SelectedProfile.GroupProbeTargets) == 0 {
		return "消费组=未配置"
	}
	if bundle == nil || bundle.Group == nil {
		return "消费组=缺失"
	}
	return fmt.Sprintf("消费组=已采集(%d组)", len(bundle.Group.Targets))
}

func summarizeJMXCoverage(env *config.Runtime, bundle *snapshot.Bundle) string {
	if !env.EnableJMX {
		return "JMX=未启用"
	}
	if bundle == nil || bundle.Metrics == nil {
		return "JMX=缺失"
	}
	if bundle.Metrics.Available {
		return fmt.Sprintf("JMX=已采集(%d个端点)", len(bundle.Metrics.Endpoints))
	}
	return "JMX=未采到可用指标"
}

func summarizeDockerCoverage(env *config.Runtime, bundle *snapshot.Bundle) string {
	if !env.EnableDocker {
		return "Docker=未启用"
	}
	return fmt.Sprintf("Docker=%s", snapshotState(bundle != nil && bundle.Docker != nil, "已采集", "缺失"))
}

func summarizeHostCoverage(env *config.Runtime, bundle *snapshot.Bundle) string {
	if !env.EnableHost {
		return "宿主机=未启用"
	}
	return fmt.Sprintf("宿主机=%s", snapshotState(bundle != nil && bundle.Host != nil, "已采集", "缺失"))
}

func summarizeLogCoverage(env *config.Runtime, bundle *snapshot.Bundle) string {
	if !env.Config.Logs.Enabled {
		return "日志=未启用"
	}
	return fmt.Sprintf("日志=%s", snapshotState(bundle != nil && bundle.Logs != nil, "已采集", "缺失"))
}

func summarizeProbeCoverage(env *config.Runtime, bundle *snapshot.Bundle) string {
	if bundle == nil || bundle.Probe == nil {
		return "探针=缺失"
	}
	if bundle.Probe.Skipped {
		return fmt.Sprintf("探针=跳过(%s)", bundle.Probe.Reason)
	}
	if !shouldRunProbeByMode(env.Mode, env.Config.Probe.Enabled) {
		return "探针=未启用"
	}
	return "探针=已执行"
}

func shouldRunProbeByMode(mode string, enabled bool) bool {
	if !enabled {
		return false
	}
	switch mode {
	case "probe", "full", "incident":
		return true
	default:
		return false
	}
}

func snapshotState(ok bool, yes string, no string) string {
	if ok {
		return yes
	}
	return no
}
