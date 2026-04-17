package logs

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type FingerprintChecker struct{}

func (FingerprintChecker) ID() string     { return "LOG-002" }
func (FingerprintChecker) Name() string   { return "error_fingerprints" }
func (FingerprintChecker) Module() string { return "logs" }

func (FingerprintChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	logs := logSnap(bundle)
	if logs == nil || !logs.Collected || len(logs.Sources) == 0 {
		return rule.NewSkip("LOG-002", "error_fingerprints", "logs", "log fingerprints cannot be evaluated without collected log sources")
	}
	if len(logs.Matches) == 0 {
		return rule.NewPass("LOG-002", "error_fingerprints", "logs", "no known Kafka error fingerprints were found in recent logs")
	}

	result := resultForMatchSeverity("LOG-002", "error_fingerprints", "known Kafka error fingerprints were found in recent logs", highestSeverity(logs.Matches))
	for _, match := range logs.Matches {
		result.Evidence = append(result.Evidence, fmt.Sprintf("%s severity=%s count=%d sources=%v", match.ID, match.Severity, match.Count, match.AffectedSources))
		if len(result.PossibleCauses) < 4 {
			result.PossibleCauses = append(result.PossibleCauses, match.ProbableCauses...)
		}
		if len(result.NextActions) < 4 {
			result.NextActions = append(result.NextActions, match.RecommendedChecks...)
		}
	}
	return result
}
