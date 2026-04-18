package kraft

import (
	"context"
	"fmt"
	"math"
	"sort"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type EpochChecker struct{}

func (EpochChecker) ID() string     { return "KRF-006" }
func (EpochChecker) Name() string   { return "controller_epoch_stability" }
func (EpochChecker) Module() string { return "kraft" }

func (EpochChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := metricsSkip("KRF-006", "controller_epoch_stability", "kraft", bundle); skip {
		return result
	}

	epochs := map[string]float64{}
	leaders := map[string]float64{}
	evidence := []string{}
	for _, endpoint := range bundle.Metrics.Endpoints {
		for name, value := range endpoint.Metrics {
			switch name {
			case "kafka_server_raftmanager_currentepoch":
				epochs[endpoint.Address] = value
				evidence = append(evidence, fmt.Sprintf("endpoint=%s current_epoch=%.0f", endpoint.Address, value))
			case "kafka_server_raftmanager_currentleader":
				leaders[endpoint.Address] = value
				evidence = append(evidence, fmt.Sprintf("endpoint=%s current_leader=%.0f", endpoint.Address, value))
			}
		}
	}
	if len(epochs) == 0 && len(leaders) == 0 {
		return rule.NewSkip("KRF-006", "controller_epoch_stability", "kraft", "current epoch metrics are not available in the current JMX sources")
	}

	result := rule.NewPass("KRF-006", "controller_epoch_stability", "kraft", "controller epoch and leader view are stable across the current JMX endpoints")
	result.Evidence = evidence

	if len(uniqueRoundedValues(epochs)) > 1 || len(uniqueRoundedValues(leaders)) > 1 {
		result = rule.NewWarn("KRF-006", "controller_epoch_stability", "kraft", "different JMX endpoints report different controller epoch or leader values")
		result.Evidence = evidence
		result.NextActions = []string{"check whether controller election is still converging", "inspect controller logs for repeated leader changes", "compare the current controller view from each broker host"}
		return result
	}

	if spread := valueSpread(epochs); spread > 1 {
		result = rule.NewWarn("KRF-006", "controller_epoch_stability", "kraft", "controller epoch spread across brokers is larger than expected and may indicate recent election churn")
		result.Evidence = append(result.Evidence, fmt.Sprintf("epoch_spread=%.0f", spread))
		result.NextActions = []string{"inspect recent controller election history", "check whether controller endpoints experienced intermittent failures", "watch current epoch over a short window to confirm whether the value keeps increasing"}
	}
	return result
}

func uniqueRoundedValues(values map[string]float64) []int64 {
	out := make([]int64, 0, len(values))
	seen := map[int64]struct{}{}
	for _, value := range values {
		rounded := int64(math.Round(value))
		if _, ok := seen[rounded]; ok {
			continue
		}
		seen[rounded] = struct{}{}
		out = append(out, rounded)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func valueSpread(values map[string]float64) float64 {
	if len(values) == 0 {
		return 0
	}
	minValue := math.MaxFloat64
	maxValue := -math.MaxFloat64
	for _, value := range values {
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue - minValue
}
