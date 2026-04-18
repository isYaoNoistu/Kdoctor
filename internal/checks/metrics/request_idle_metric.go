package metrics

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type RequestIdleMetricChecker struct {
	Warn float64
	Crit float64
}

func (RequestIdleMetricChecker) ID() string     { return "MET-006" }
func (RequestIdleMetricChecker) Name() string   { return "request_idle_metric" }
func (RequestIdleMetricChecker) Module() string { return "metrics" }

func (c RequestIdleMetricChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("MET-006", "request_idle_metric", "metrics", bundle); skip {
		return result
	}
	if c.Warn <= 0 {
		c.Warn = 0.3
	}
	if c.Crit <= 0 {
		c.Crit = 0.1
	}

	value, ok, evidence := aggregateMin(metricsSnap(bundle), "kafka_server_kafkarequesthandlerpool_requesthandleravgidlepercent")
	if !ok {
		return rule.NewSkip("MET-006", "request_idle_metric", "metrics", "request handler idle metrics are not available in the current JMX sources")
	}
	evidence = append(evidence, fmt.Sprintf("request_idle_min=%.3f", value))

	result := rule.NewPass("MET-006", "request_idle_metric", "metrics", "request handler idle metrics still show healthy processing headroom")
	result.Evidence = evidence
	switch {
	case value <= c.Crit:
		result = rule.NewFail("MET-006", "request_idle_metric", "metrics", "request handler idle metrics are critically low and broker request threads are close to saturation")
		result.Evidence = evidence
		result.NextActions = []string{"correlate request idle with queue time, purgatory, and replica pressure", "check whether recent traffic or rebalance storms are saturating broker handlers", "review disk, ISR, and GC pressure before scaling or rerouting traffic"}
	case value <= c.Warn:
		result = rule.NewWarn("MET-006", "request_idle_metric", "metrics", "request handler idle metrics are getting low and broker processing headroom is shrinking")
		result.Evidence = evidence
		result.NextActions = []string{"watch request idle over the next peak window", "correlate handler idle with request latency and quota pressure", "review whether hot partitions or hot brokers are concentrating request load"}
	}
	return result
}
