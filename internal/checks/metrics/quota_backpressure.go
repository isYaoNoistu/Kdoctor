package metrics

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type BackpressureChecker struct {
	RequestLatencyWarnMs float64
}

func (BackpressureChecker) ID() string     { return "QTA-004" }
func (BackpressureChecker) Name() string   { return "broker_backpressure" }
func (BackpressureChecker) Module() string { return "quota" }

func (c BackpressureChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("QTA-004", "broker_backpressure", "quota", bundle); skip {
		return result
	}
	if c.RequestLatencyWarnMs <= 0 {
		c.RequestLatencyWarnMs = 100
	}

	idle, idleOK, idleEvidence := aggregateMinMatching(metricsSnap(bundle), func(name string) bool {
		return containsAllMetric(name, "idlepercent")
	})
	latency, latencyOK, latencyEvidence := aggregateMaxMatching(metricsSnap(bundle), func(name string) bool {
		return containsAllMetric(name, "request", "time")
	})
	if !idleOK && !latencyOK {
		return rule.NewSkip("QTA-004", "broker_backpressure", "quota", "backpressure-related idle or request latency metrics are not available in the current JMX sources")
	}

	evidence := append([]string{}, idleEvidence...)
	evidence = append(evidence, latencyEvidence...)
	result := rule.NewPass("QTA-004", "broker_backpressure", "quota", "current JMX metrics do not show a strong backpressure signal")
	result.Evidence = evidence
	if (idleOK && idle < 0.2) || (latencyOK && latency >= c.RequestLatencyWarnMs) {
		result = rule.NewWarn("QTA-004", "broker_backpressure", "quota", "broker idle headroom or request latency already suggests backpressure")
		result.Evidence = append(result.Evidence, fmt.Sprintf("min_idle=%.3f max_request_metric=%.3f", idle, latency))
		result.NextActions = []string{"correlate backpressure with quotas, network idle, and request handler idle", "check whether the pressure comes from throttling, hot partitions, or a genuine broker overload", "avoid attributing the slowdown to pure network issues before reviewing broker metrics"}
	}
	return result
}
