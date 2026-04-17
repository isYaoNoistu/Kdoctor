package logs

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ExplanationChecker struct{}

func (ExplanationChecker) ID() string     { return "LOG-003" }
func (ExplanationChecker) Name() string   { return "error_explanations" }
func (ExplanationChecker) Module() string { return "logs" }

func (ExplanationChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	logs := logSnap(bundle)
	if logs == nil || !logs.Collected || len(logs.Sources) == 0 {
		return rule.NewSkip("LOG-003", "error_explanations", "logs", "log explanations cannot be generated without collected log sources")
	}
	if len(logs.Matches) == 0 {
		return rule.NewPass("LOG-003", "error_explanations", "logs", "no matched log errors required explanation")
	}

	result := rule.NewWarn("LOG-003", "error_explanations", "logs", "matched log errors were explained and mapped to likely causes")
	for _, match := range logs.Matches {
		result.Evidence = append(result.Evidence, fmt.Sprintf("%s meaning=%s", match.ID, match.Meaning))
		if len(result.PossibleCauses) < 5 {
			result.PossibleCauses = append(result.PossibleCauses, match.ProbableCauses...)
		}
		if len(result.NextActions) < 5 {
			result.NextActions = append(result.NextActions, match.RecommendedChecks...)
		}
	}
	return result
}
