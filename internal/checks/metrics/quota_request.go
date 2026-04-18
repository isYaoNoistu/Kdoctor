package metrics

import (
	"context"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type RequestQuotaChecker struct{}

func (RequestQuotaChecker) ID() string     { return "QTA-003" }
func (RequestQuotaChecker) Name() string   { return "request_percentage_quota" }
func (RequestQuotaChecker) Module() string { return "quota" }

func (RequestQuotaChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("QTA-003", "request_percentage_quota", "quota", bundle); skip {
		return result
	}

	value, ok, evidence := aggregateMaxMatching(metricsSnap(bundle), func(name string) bool {
		return containsAllMetric(name, "request", "percentage")
	})
	if !ok {
		return rule.NewSkip("QTA-003", "request_percentage_quota", "quota", "request percentage quota metrics are not available in the current JMX sources")
	}

	result := rule.NewPass("QTA-003", "request_percentage_quota", "quota", "request percentage quota usage looks healthy in the current JMX window")
	result.Evidence = evidence
	switch {
	case value >= 1:
		result = rule.NewWarn("QTA-003", "request_percentage_quota", "quota", "request percentage quota is saturated and may already be throttling client requests")
		result.Evidence = evidence
		result.NextActions = []string{"review request percentage quotas for the affected client or tenant", "correlate quota saturation with producer or fetch throttle metrics", "decide whether quota tuning or traffic shaping is the safer mitigation"}
	case value >= 0.8:
		result = rule.NewWarn("QTA-003", "request_percentage_quota", "quota", "request percentage quota usage is close to saturation")
		result.Evidence = evidence
		result.NextActions = []string{"watch the request quota headroom during peak traffic", "check whether a single tenant or workflow is driving the rise", "tune quotas only after confirming the current ceiling is intentional"}
	}
	return result
}
