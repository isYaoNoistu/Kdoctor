package host

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type FDChecker struct {
	WarnPct int
	CritPct int
}

func (FDChecker) ID() string     { return "HOST-008" }
func (FDChecker) Name() string   { return "fd_headroom" }
func (FDChecker) Module() string { return "host" }

func (c FDChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Host == nil || !bundle.Host.Collected || bundle.Host.FD == nil {
		return rule.NewSkip("HOST-008", "fd_headroom", "host", "file descriptor evidence is not available in the current input mode")
	}
	if c.WarnPct <= 0 {
		c.WarnPct = 70
	}
	if c.CritPct <= 0 {
		c.CritPct = 85
	}

	fd := bundle.Host.FD
	evidence := []string{}
	if fd.SoftLimit > 0 {
		evidence = append(evidence, fmt.Sprintf("soft_limit=%d", fd.SoftLimit))
	}
	if fd.SystemMax > 0 {
		usedPct := float64(fd.SystemUsed) * 100 / float64(fd.SystemMax)
		evidence = append(evidence, fmt.Sprintf("system_used=%d system_max=%d used_pct=%.1f", fd.SystemUsed, fd.SystemMax, usedPct))
		result := rule.NewPass("HOST-008", "fd_headroom", "host", "host file descriptor headroom looks acceptable")
		result.Evidence = evidence
		switch {
		case usedPct >= float64(c.CritPct) || (fd.SoftLimit > 0 && fd.SoftLimit < 32768):
			result = rule.NewFail("HOST-008", "fd_headroom", "host", "host file descriptor headroom is critically low")
			result.Evidence = evidence
			result.NextActions = []string{"raise ulimit -n for Kafka and the execution environment", "inspect current file descriptor growth and socket churn", "verify recent connection spikes did not exhaust shared host limits"}
		case usedPct >= float64(c.WarnPct) || (fd.SoftLimit > 0 && fd.SoftLimit < 65536):
			result = rule.NewWarn("HOST-008", "fd_headroom", "host", "host file descriptor headroom is getting tight")
			result.Evidence = evidence
			result.NextActions = []string{"review ulimit -n and current descriptor pressure before traffic increases", "check whether connection churn or client retries are inflating descriptor usage", "reserve more fd headroom for Kafka data and network workloads"}
		}
		return result
	}

	result := rule.NewPass("HOST-008", "fd_headroom", "host", "host file descriptor soft limit is visible and does not show an immediate risk")
	result.Evidence = evidence
	if fd.SoftLimit > 0 && fd.SoftLimit < 65536 {
		result = rule.NewWarn("HOST-008", "fd_headroom", "host", "host file descriptor soft limit is lower than the typical Kafka production baseline")
		result.Evidence = evidence
		result.NextActions = []string{"raise ulimit -n for the Kafka service user", "confirm the broker process inherits the intended soft and hard limits", "review listener and client connection fan-out before load grows"}
	}
	return result
}
