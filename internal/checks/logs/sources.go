package logs

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type SourcesChecker struct{}

func (SourcesChecker) ID() string     { return "LOG-001" }
func (SourcesChecker) Name() string   { return "log_sources" }
func (SourcesChecker) Module() string { return "logs" }

func (SourcesChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	logs := logSnap(bundle)
	if logs == nil || !logs.Collected {
		return rule.NewSkip("LOG-001", "log_sources", "logs", "log collection is not enabled in the current input mode")
	}
	if len(logs.Sources) == 0 {
		result := rule.NewSkip("LOG-001", "log_sources", "logs", "no log sources were available from the current execution view")
		result.Evidence = append(result.Evidence, logs.Errors...)
		return result
	}

	result := rule.NewPass("LOG-001", "log_sources", "logs", "log sources were collected successfully")
	for _, source := range logs.Sources {
		result.Evidence = append(result.Evidence, fmt.Sprintf("source=%s", source))
	}
	for _, err := range logs.Errors {
		result.Evidence = append(result.Evidence, fmt.Sprintf("collector warning=%s", err))
	}
	return result
}
