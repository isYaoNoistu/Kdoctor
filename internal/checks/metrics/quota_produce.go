package metrics

import (
	"context"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ProduceThrottleChecker struct {
	WarnMs float64
}

func (ProduceThrottleChecker) ID() string     { return "QTA-001" }
func (ProduceThrottleChecker) Name() string   { return "produce_throttle" }
func (ProduceThrottleChecker) Module() string { return "quota" }

func (c ProduceThrottleChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("QTA-001", "produce_throttle", "quota", bundle); skip {
		return result
	}
	if c.WarnMs <= 0 {
		c.WarnMs = 1
	}

	value, ok, evidence := aggregateMaxMatching(metricsSnap(bundle), func(name string) bool {
		return containsAllMetric(name, "throttle", "produce")
	})
	if !ok {
		return rule.NewSkip("QTA-001", "produce_throttle", "quota", "produce throttle metrics are not available in the current JMX sources")
	}

	result := rule.NewPass("QTA-001", "produce_throttle", "quota", "produce throttle time is not visible in the current JMX window")
	result.Evidence = evidence
	if value >= c.WarnMs {
		result = rule.NewWarn("QTA-001", "produce_throttle", "quota", "produce throttle time is above zero and may already be limiting write throughput")
		result.Evidence = evidence
		result.NextActions = []string{"review produce quotas or tenant limits", "compare throttle time with producer latency and idle metrics", "separate quota pressure from broker instability before changing retries or timeouts"}
	}
	return result
}
