package producer

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ThrottleChecker struct {
	WarnMs float64
}

func (ThrottleChecker) ID() string     { return "PRD-005" }
func (ThrottleChecker) Name() string   { return "producer_throttle" }
func (ThrottleChecker) Module() string { return "producer" }

func (c ThrottleChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Metrics == nil || !bundle.Metrics.Collected {
		return rule.NewSkip("PRD-005", "producer_throttle", "producer", "JMX metrics are not enabled in the current input mode")
	}
	if !bundle.Metrics.Available {
		result := rule.NewSkip("PRD-005", "producer_throttle", "producer", "no usable JMX metrics source is available")
		result.Evidence = append(result.Evidence, bundle.Metrics.Errors...)
		return result
	}

	if c.WarnMs <= 0 {
		c.WarnMs = 1
	}
	value, ok, evidence := producerThrottle(bundle.Metrics)
	if !ok {
		return rule.NewSkip("PRD-005", "producer_throttle", "producer", "producer throttle metrics are not available in the current JMX sources")
	}

	result := rule.NewPass("PRD-005", "producer_throttle", "producer", "producer throttle metrics do not currently show quota pressure")
	result.Evidence = evidence
	if value >= c.WarnMs {
		result = rule.NewWarn("PRD-005", "producer_throttle", "producer", "producer throttle time is above zero and may already be affecting write latency")
		result.Evidence = evidence
		result.NextActions = []string{"review producer quota settings and tenant limits", "compare throttle time with request latency and broker idle metrics", "distinguish quota pressure from broker availability problems before tuning retries"}
	}
	return result
}

func producerThrottle(metrics *snapshot.MetricsSnapshot) (float64, bool, []string) {
	best := 0.0
	found := false
	evidence := []string{}
	for _, endpoint := range metrics.Endpoints {
		for name, value := range endpoint.Metrics {
			lower := strings.ToLower(name)
			if !strings.Contains(lower, "throttle") || !strings.Contains(lower, "produce") {
				continue
			}
			if !found || value > best {
				best = value
			}
			found = true
			evidence = append(evidence, fmt.Sprintf("endpoint=%s metric=%s value=%.3f", endpoint.Address, name, value))
		}
	}
	return best, found, evidence
}
