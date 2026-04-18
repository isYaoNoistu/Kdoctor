package metrics

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type RequestPressureChecker struct {
	WarnLatencyMs float64
	WarnPurgatory float64
}

func (RequestPressureChecker) ID() string     { return "JVM-003" }
func (RequestPressureChecker) Name() string   { return "request_latency_and_purgatory" }
func (RequestPressureChecker) Module() string { return "jvm" }

func (c RequestPressureChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("JVM-003", "request_latency_and_purgatory", "jvm", bundle); skip {
		return result
	}
	if c.WarnLatencyMs <= 0 {
		c.WarnLatencyMs = 100
	}
	if c.WarnPurgatory <= 0 {
		c.WarnPurgatory = 1
	}

	latency, latencyOK, latencyEvidence := aggregateMaxMatching(metricsSnap(bundle), func(name string) bool {
		return containsAllMetric(name, "request", "time")
	})
	purgatory, purgatoryOK, purgatoryEvidence := aggregateMaxMatching(metricsSnap(bundle), func(name string) bool {
		return containsAllMetric(name, "purgatory")
	})
	if !latencyOK && !purgatoryOK {
		return rule.NewSkip("JVM-003", "request_latency_and_purgatory", "jvm", "request latency or purgatory metrics are not available in the current JMX sources")
	}

	evidence := append([]string{}, latencyEvidence...)
	evidence = append(evidence, purgatoryEvidence...)
	result := rule.NewPass("JVM-003", "request_latency_and_purgatory", "jvm", "request latency and purgatory metrics do not currently show queueing pressure")
	result.Evidence = evidence
	if (latencyOK && latency >= c.WarnLatencyMs) || (purgatoryOK && purgatory >= c.WarnPurgatory) {
		result = rule.NewWarn("JVM-003", "request_latency_and_purgatory", "jvm", "request latency or purgatory backlog is elevated and may already be contributing to broker pressure")
		result.Evidence = append(result.Evidence, fmt.Sprintf("max_request_metric=%.3f max_purgatory=%.3f", latency, purgatory))
		result.NextActions = []string{"check whether the latency comes from network, disk, replication, or quota pressure", "correlate request queue metrics with handler idle and ISR status", "inspect recent traffic spikes before changing broker thread settings"}
	}
	return result
}
