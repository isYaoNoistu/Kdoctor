package metrics

import (
	"context"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type FetchThrottleChecker struct {
	WarnMs float64
}

func (FetchThrottleChecker) ID() string     { return "QTA-002" }
func (FetchThrottleChecker) Name() string   { return "fetch_throttle" }
func (FetchThrottleChecker) Module() string { return "quota" }

func (c FetchThrottleChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("QTA-002", "fetch_throttle", "quota", bundle); skip {
		return result
	}
	if c.WarnMs <= 0 {
		c.WarnMs = 1
	}

	value, ok, evidence := aggregateMaxMatching(metricsSnap(bundle), func(name string) bool {
		return containsAllMetric(name, "throttle", "fetch")
	})
	if !ok {
		return rule.NewSkip("QTA-002", "fetch_throttle", "quota", "fetch throttle metrics are not available in the current JMX sources")
	}

	result := rule.NewPass("QTA-002", "fetch_throttle", "quota", "fetch throttle time is not visible in the current JMX window")
	result.Evidence = evidence
	if value >= c.WarnMs {
		result = rule.NewWarn("QTA-002", "fetch_throttle", "quota", "fetch throttle time is above zero and may already be slowing consumers")
		result.Evidence = evidence
		result.NextActions = []string{"review fetch quotas or tenant limits", "compare throttle time with consumer lag and coordinator health", "distinguish quota-induced slowness from partition or ISR problems"}
	}
	return result
}
