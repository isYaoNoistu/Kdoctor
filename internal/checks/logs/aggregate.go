package logs

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type AggregateChecker struct{}

func (AggregateChecker) ID() string     { return "LOG-004" }
func (AggregateChecker) Name() string   { return "duplicate_aggregation" }
func (AggregateChecker) Module() string { return "logs" }

func (AggregateChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	logs := logSnap(bundle)
	if logs == nil || !logs.Collected || len(logs.Sources) == 0 {
		return rule.NewSkip("LOG-004", "duplicate_aggregation", "logs", "log aggregation cannot be evaluated without collected log sources")
	}
	if len(logs.Matches) == 0 {
		return rule.NewPass("LOG-004", "duplicate_aggregation", "logs", "no repeated log fingerprints were observed")
	}

	result := rule.NewPass("LOG-004", "duplicate_aggregation", "logs", "repeated log fingerprints were aggregated successfully")
	for _, match := range logs.Matches {
		result.Evidence = append(result.Evidence, fmt.Sprintf("%s count=%d sources=%d", match.ID, match.Count, len(match.AffectedSources)))
	}
	return result
}
