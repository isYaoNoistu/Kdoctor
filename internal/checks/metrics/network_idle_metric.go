package metrics

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type NetworkIdleMetricChecker struct {
	Warn float64
	Crit float64
}

func (NetworkIdleMetricChecker) ID() string     { return "MET-005" }
func (NetworkIdleMetricChecker) Name() string   { return "network_idle_metric" }
func (NetworkIdleMetricChecker) Module() string { return "metrics" }

func (c NetworkIdleMetricChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("MET-005", "network_idle_metric", "metrics", bundle); skip {
		return result
	}
	if c.Warn <= 0 {
		c.Warn = 0.3
	}
	if c.Crit <= 0 {
		c.Crit = 0.1
	}

	value, ok, evidence := aggregateMin(metricsSnap(bundle), "kafka_network_socketserver_networkprocessoravgidlepercent")
	if !ok {
		return rule.NewSkip("MET-005", "network_idle_metric", "metrics", "network idle metrics are not available in the current JMX sources")
	}
	evidence = append(evidence, fmt.Sprintf("network_idle_min=%.3f", value))

	result := rule.NewPass("MET-005", "network_idle_metric", "metrics", "network idle metrics still show healthy headroom")
	result.Evidence = evidence
	switch {
	case value <= c.Crit:
		result = rule.NewFail("MET-005", "network_idle_metric", "metrics", "network idle metrics are critically low and already suggest broker-side saturation")
		result.Evidence = evidence
		result.NextActions = []string{"check connection churn and listener traffic concentration", "correlate network idle with request latency and quota/backpressure signals", "verify the current route design is not funneling all traffic through a hot broker"}
	case value <= c.Warn:
		result = rule.NewWarn("MET-005", "network_idle_metric", "metrics", "network idle metrics are getting low and broker network headroom is shrinking")
		result.Evidence = evidence
		result.NextActions = []string{"watch network idle over a longer window", "check whether the current traffic spike or client fan-out is sustainable", "review listener routing and load distribution before pressure worsens"}
	}
	return result
}
