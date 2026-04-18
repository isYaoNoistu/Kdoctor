package metrics

import (
	"fmt"
	"math"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func metricsSnap(bundle *snapshot.Bundle) *snapshot.MetricsSnapshot {
	if bundle == nil {
		return nil
	}
	return bundle.Metrics
}

func skipIfUnavailable(id string, name string, module string, bundle *snapshot.Bundle) (model.CheckResult, bool) {
	metrics := metricsSnap(bundle)
	if metrics == nil || !metrics.Collected {
		return rule.NewSkip(id, name, module, "当前输入模式未启用 JMX 指标采集"), true
	}
	if !metrics.Available {
		result := rule.NewSkip(id, name, module, "当前没有可用的 JMX 指标来源")
		result.Evidence = append(result.Evidence, metrics.Errors...)
		return result, true
	}
	return model.CheckResult{}, false
}

func aggregateMax(metrics *snapshot.MetricsSnapshot, names ...string) (float64, bool, []string) {
	best := 0.0
	found := false
	evidence := []string{}
	for _, endpoint := range metrics.Endpoints {
		for _, name := range names {
			value, ok := endpoint.Metrics[strings.ToLower(name)]
			if !ok {
				continue
			}
			if !found || value > best {
				best = value
			}
			found = true
			evidence = append(evidence, fmt.Sprintf("端点=%s 指标=%s 值=%.3f", endpoint.Address, name, value))
		}
	}
	return best, found, evidence
}

func aggregateMin(metrics *snapshot.MetricsSnapshot, names ...string) (float64, bool, []string) {
	best := math.MaxFloat64
	found := false
	evidence := []string{}
	for _, endpoint := range metrics.Endpoints {
		for _, name := range names {
			value, ok := endpoint.Metrics[strings.ToLower(name)]
			if !ok {
				continue
			}
			if !found || value < best {
				best = value
			}
			found = true
			evidence = append(evidence, fmt.Sprintf("端点=%s 指标=%s 值=%.3f", endpoint.Address, name, value))
		}
	}
	if !found {
		return 0, false, evidence
	}
	return best, true, evidence
}

func aggregateMaxMatching(metrics *snapshot.MetricsSnapshot, match func(string) bool) (float64, bool, []string) {
	best := 0.0
	found := false
	evidence := []string{}
	for _, endpoint := range metrics.Endpoints {
		for name, value := range endpoint.Metrics {
			if !match(strings.ToLower(name)) {
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

func aggregateMinMatching(metrics *snapshot.MetricsSnapshot, match func(string) bool) (float64, bool, []string) {
	best := math.MaxFloat64
	found := false
	evidence := []string{}
	for _, endpoint := range metrics.Endpoints {
		for name, value := range endpoint.Metrics {
			if !match(strings.ToLower(name)) {
				continue
			}
			if !found || value < best {
				best = value
			}
			found = true
			evidence = append(evidence, fmt.Sprintf("endpoint=%s metric=%s value=%.3f", endpoint.Address, name, value))
		}
	}
	if !found {
		return 0, false, evidence
	}
	return best, true, evidence
}

func distinctMetricValues(metrics *snapshot.MetricsSnapshot, match func(string) bool) (map[string]float64, bool, []string) {
	values := map[string]float64{}
	evidence := []string{}
	for _, endpoint := range metrics.Endpoints {
		for name, value := range endpoint.Metrics {
			if !match(strings.ToLower(name)) {
				continue
			}
			values[fmt.Sprintf("%s|%s", endpoint.Address, name)] = value
			evidence = append(evidence, fmt.Sprintf("endpoint=%s metric=%s value=%.3f", endpoint.Address, name, value))
		}
	}
	return values, len(values) > 0, evidence
}

func containsAllMetric(name string, fragments ...string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, fragment := range fragments {
		if !strings.Contains(name, strings.ToLower(strings.TrimSpace(fragment))) {
			return false
		}
	}
	return true
}
