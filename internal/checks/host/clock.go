package host

import (
	"context"
	"fmt"
	"math"
	"time"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ClockChecker struct {
	WarnMs int
}

func (ClockChecker) ID() string     { return "HOST-009" }
func (ClockChecker) Name() string   { return "clock_skew" }
func (ClockChecker) Module() string { return "host" }

func (c ClockChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Metrics == nil || !bundle.Metrics.Collected {
		return rule.NewSkip("HOST-009", "clock_skew", "host", "clock skew evidence is not available because JMX metrics are disabled")
	}
	if !bundle.Metrics.Available {
		return rule.NewSkip("HOST-009", "clock_skew", "host", "clock skew evidence is not available because no JMX endpoint responded")
	}
	if c.WarnMs <= 0 {
		c.WarnMs = 500
	}

	now := time.Now().Unix()
	samples := []int64{}
	evidence := []string{}
	for _, endpoint := range bundle.Metrics.Endpoints {
		if endpoint.ServerTimeUnix == 0 {
			continue
		}
		samples = append(samples, endpoint.ServerTimeUnix)
		deltaMs := math.Abs(float64(endpoint.ServerTimeUnix-now)) * 1000
		evidence = append(evidence, fmt.Sprintf("endpoint=%s server_time_unix=%d delta_ms=%.0f", endpoint.Address, endpoint.ServerTimeUnix, deltaMs))
	}
	if len(samples) == 0 {
		return rule.NewSkip("HOST-009", "clock_skew", "host", "JMX endpoints did not return usable server time headers")
	}

	maxSkewMs := maxPairwiseSkewMs(samples)
	result := rule.NewPass("HOST-009", "clock_skew", "host", "clock skew between the current host and JMX endpoints is within the warning window")
	result.Evidence = append(evidence, fmt.Sprintf("max_pairwise_skew_ms=%.0f", maxSkewMs))
	if maxSkewMs >= float64(c.WarnMs) {
		result = rule.NewWarn("HOST-009", "clock_skew", "host", "clock skew between hosts is larger than expected and can affect SSL, logs, and transaction timing")
		result.Evidence = append(evidence, fmt.Sprintf("max_pairwise_skew_ms=%.0f", maxSkewMs))
		result.NextActions = []string{"verify NTP or chrony status on the broker hosts", "compare system time on controller and broker nodes directly", "treat log ordering and certificate validation cautiously until clocks are aligned"}
	}
	return result
}

func maxPairwiseSkewMs(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}
	minValue := values[0]
	maxValue := values[0]
	for _, value := range values[1:] {
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}
	return math.Abs(float64(maxValue-minValue)) * 1000
}
