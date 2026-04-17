package lint

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ComposeChecker struct{}

func (ComposeChecker) ID() string     { return "CFG-001" }
func (ComposeChecker) Name() string   { return "compose_parse" }
func (ComposeChecker) Module() string { return "lint" }

func (ComposeChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Compose == nil {
		return rule.NewSkip("CFG-001", "compose_parse", "lint", "compose snapshot not available")
	}

	services := kafkaServices(snap.Compose)
	result := rule.NewPass("CFG-001", "compose_parse", "lint", "compose parsed successfully")
	result.Evidence = []string{
		fmt.Sprintf("source=%s", snap.Compose.SourcePath),
		fmt.Sprintf("kafka_services=%d", len(services)),
	}
	if len(services) == 0 {
		result = rule.NewWarn("CFG-001", "compose_parse", "lint", "compose parsed but no Kafka services were detected")
		result.Evidence = []string{fmt.Sprintf("source=%s", snap.Compose.SourcePath)}
	}
	return result
}
