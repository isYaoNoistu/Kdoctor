package metrics

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type HeapGCChecker struct {
	HeapWarnPct   float64
	GCPauseWarnMs float64
}

func (HeapGCChecker) ID() string     { return "JVM-004" }
func (HeapGCChecker) Name() string   { return "heap_and_gc_pressure" }
func (HeapGCChecker) Module() string { return "jvm" }

func (c HeapGCChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("JVM-004", "heap_and_gc_pressure", "jvm", bundle); skip {
		return result
	}
	if c.HeapWarnPct <= 0 {
		c.HeapWarnPct = 85
	}
	if c.GCPauseWarnMs <= 0 {
		c.GCPauseWarnMs = 200
	}

	heapPct, heapOK, heapEvidence := heapUsedPercent(metricsSnap(bundle))
	gcPause, gcOK, gcEvidence := aggregateMaxMatching(metricsSnap(bundle), func(name string) bool {
		return strings.Contains(name, "gc") && (strings.Contains(name, "pause") || strings.Contains(name, "duration") || strings.Contains(name, "collectiontime"))
	})
	if !heapOK && !gcOK {
		return rule.NewSkip("JVM-004", "heap_and_gc_pressure", "jvm", "heap or GC pause metrics are not available in the current JMX sources")
	}

	evidence := append([]string{}, heapEvidence...)
	evidence = append(evidence, gcEvidence...)
	result := rule.NewPass("JVM-004", "heap_and_gc_pressure", "jvm", "heap usage and GC pause metrics do not currently show strong JVM pressure")
	result.Evidence = evidence
	if (heapOK && heapPct >= c.HeapWarnPct) || (gcOK && gcPause >= c.GCPauseWarnMs) {
		result = rule.NewWarn("JVM-004", "heap_and_gc_pressure", "jvm", "heap usage or GC pause metrics indicate rising JVM pressure")
		result.Evidence = append(result.Evidence, fmt.Sprintf("heap_used_pct=%.2f max_gc_pause_metric=%.3f", heapPct, gcPause))
		result.NextActions = []string{"review JVM heap sizing and container memory headroom together", "check whether request backlogs or large batches are amplifying heap pressure", "inspect GC logs or JMX trends before increasing heap blindly"}
	}
	return result
}

func heapUsedPercent(metrics *snapshot.MetricsSnapshot) (float64, bool, []string) {
	if metrics == nil {
		return 0, false, nil
	}

	used := 0.0
	usedFound := false
	maxValue := 0.0
	maxFound := false
	evidence := []string{}
	percent, percentOK, percentEvidence := aggregateMaxMatching(metrics, func(name string) bool {
		return strings.Contains(name, "heap") && strings.Contains(name, "usedpercent")
	})
	if percentOK {
		return percent, true, percentEvidence
	}

	for _, endpoint := range metrics.Endpoints {
		for name, value := range endpoint.Metrics {
			lower := strings.ToLower(name)
			switch {
			case strings.Contains(lower, "heapmemoryusage_used"):
				used = value
				usedFound = true
				evidence = append(evidence, fmt.Sprintf("endpoint=%s metric=%s value=%.0f", endpoint.Address, name, value))
			case strings.Contains(lower, "heapmemoryusage_max"):
				maxValue = value
				maxFound = true
				evidence = append(evidence, fmt.Sprintf("endpoint=%s metric=%s value=%.0f", endpoint.Address, name, value))
			}
		}
	}

	if !usedFound || !maxFound || maxValue <= 0 {
		return 0, false, evidence
	}
	return used * 100 / maxValue, true, evidence
}
