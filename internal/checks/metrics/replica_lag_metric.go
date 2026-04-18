package metrics

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ReplicaLagChecker struct {
	Warn int64
}

func (ReplicaLagChecker) ID() string     { return "MET-003" }
func (ReplicaLagChecker) Name() string   { return "replica_fetcher_lag" }
func (ReplicaLagChecker) Module() string { return "metrics" }

func (c ReplicaLagChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if result, skip := skipIfUnavailable("MET-003", "replica_fetcher_lag", "metrics", bundle); skip {
		return result
	}
	if c.Warn <= 0 {
		c.Warn = 10000
	}

	value, ok, evidence := aggregateMaxMatching(metricsSnap(bundle), func(name string) bool {
		name = strings.ToLower(name)
		if !strings.Contains(name, "replicafetchermanager") {
			return false
		}
		return strings.Contains(name, "maxlag") || strings.Contains(name, "consumerlag")
	})
	if !ok {
		return rule.NewSkip("MET-003", "replica_fetcher_lag", "metrics", "replica lag metrics are not available in the current JMX sources")
	}
	evidence = append(evidence, fmt.Sprintf("replica_lag_max=%.0f", value))

	result := rule.NewPass("MET-003", "replica_fetcher_lag", "metrics", "replica fetcher lag metrics look healthy in the current JMX window")
	result.Evidence = evidence
	if int64(value) >= c.Warn {
		result = rule.NewWarn("MET-003", "replica_fetcher_lag", "metrics", "replica fetcher lag is elevated and may soon erode ISR safety")
		result.Evidence = evidence
		result.NextActions = []string{"check follower broker disk and network pressure", "correlate replica lag with ISR and under-replicated partition signals", "review whether the current write burst is outrunning replica catch-up capacity"}
	}
	return result
}
