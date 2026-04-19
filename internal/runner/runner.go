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
	coverage = append(coverage, describeCoverage("网络", networkEnabled(env), networkEvidence(bundle)))
	coverage = append(coverage, describeCoverage("Compose", composeEnabled(env), composeEvidence(bundle)))
	coverage = append(coverage, describeCoverage("Kafka", kafkaEnabled(env), kafkaEvidence(bundle)))
	coverage = append(coverage, describeCoverage("消费组", groupEnabled(env), groupEvidence(bundle)))
	coverage = append(coverage, describeCoverage("Docker", dockerEnabled(env, bundle), dockerEvidence(bundle)))
	coverage = append(coverage, describeCoverage("宿主机", hostEnabled(env, bundle), hostEvidence(bundle)))
	coverage = append(coverage, describeCoverage("日志", logEnabled(env, bundle), logEvidence(bundle)))
	coverage = append(coverage, describeCoverage("探针", probeEnabled(env), probeEvidence(env, bundle)))
	return coverage
}

func describeCoverage(name string, enabled bool, hasEvidence bool) string {
	switch {
	case !enabled:
		return fmt.Sprintf("%s=未纳入本次运行", name)
	case hasEvidence:
		return fmt.Sprintf("%s=已启用，已获取证据", name)
	default:
		return fmt.Sprintf("%s=已启用，未获取证据", name)
	}
}

func networkEnabled(env *config.Runtime) bool {
	return env != nil && (len(env.BootstrapExternal) > 0 || len(env.BootstrapInternal) > 0 || len(env.ControllerEndpoints) > 0)
}

func networkEvidence(bundle *snapshot.Bundle) bool {
	if bundle == nil || bundle.Network == nil {
		return false
	}
	return len(bundle.Network.BootstrapChecks) > 0 || len(bundle.Network.ControllerChecks) > 0 || len(bundle.Network.MetadataChecks) > 0
}

func composeEnabled(env *config.Runtime) bool {
	return env != nil && strings.TrimSpace(env.ComposePath) != ""
}

func composeEvidence(bundle *snapshot.Bundle) bool {
	return bundle != nil && bundle.Compose != nil && len(bundle.Compose.Services) > 0
}

func kafkaEnabled(env *config.Runtime) bool {
	return networkEnabled(env)
}

func kafkaEvidence(bundle *snapshot.Bundle) bool {
	return bundle != nil && bundle.Kafka != nil && len(bundle.Kafka.Brokers) > 0
}

func groupEnabled(env *config.Runtime) bool {
	return env != nil && len(env.SelectedProfile.GroupProbeTargets) > 0
}

func groupEvidence(bundle *snapshot.Bundle) bool {
	return bundle != nil && bundle.Group != nil && (len(bundle.Group.Targets) > 0 || len(bundle.Group.Errors) > 0)
}

func dockerEnabled(env *config.Runtime, bundle *snapshot.Bundle) bool {
	if env == nil || !env.EnableDocker {
		return false
	}
	if len(env.Config.Docker.ContainerNames) > 0 {
		return true
	}
	return composeEvidence(bundle)
}

func dockerEvidence(bundle *snapshot.Bundle) bool {
	if bundle == nil || bundle.Docker == nil {
		return false
	}
	return len(bundle.Docker.Containers) > 0 || len(bundle.Docker.Errors) > 0
}

func hostEnabled(env *config.Runtime, bundle *snapshot.Bundle) bool {
	if env == nil || !env.EnableHost {
		return false
	}
	if strings.TrimSpace(env.LogDir) != "" {
		return true
	}
	if len(env.Config.Host.DiskPaths) > 0 || len(env.Config.Host.CheckPorts) > 0 {
		return true
	}
	return composeEvidence(bundle) || dockerEvidence(bundle)
}

func hostEvidence(bundle *snapshot.Bundle) bool {
	if bundle == nil || bundle.Host == nil {
		return false
	}
	host := bundle.Host
	return len(host.DiskUsages) > 0 ||
		len(host.PortChecks) > 0 ||
		len(host.ObservedListenPorts) > 0 ||
		host.FD != nil ||
		host.Memory != nil
}

func logEnabled(env *config.Runtime, bundle *snapshot.Bundle) bool {
	if env == nil || !env.Config.Logs.Enabled {
		return false
	}
	if strings.TrimSpace(env.LogDir) != "" {
		return true
	}
	return dockerEnabled(env, bundle)
}

func logEvidence(bundle *snapshot.Bundle) bool {
	if bundle == nil || bundle.Logs == nil {
		return false
	}
	return len(bundle.Logs.Sources) > 0 || len(bundle.Logs.SourceStats) > 0 || len(bundle.Logs.Matches) > 0
}

func probeEnabled(env *config.Runtime) bool {
	return env != nil && shouldRunProbeByMode(env.Mode, env.Config.Probe.Enabled)
}

func probeEvidence(env *config.Runtime, bundle *snapshot.Bundle) bool {
	if !probeEnabled(env) || bundle == nil || bundle.Probe == nil {
		return false
	}
	return true
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
