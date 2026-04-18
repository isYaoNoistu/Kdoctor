package storage

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type TieredStorageChecker struct{}

func (TieredStorageChecker) ID() string     { return "STG-006" }
func (TieredStorageChecker) Name() string   { return "tiered_storage_awareness" }
func (TieredStorageChecker) Module() string { return "storage" }

func (TieredStorageChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	kafkaServices := services(bundle)
	if len(kafkaServices) == 0 {
		return rule.NewSkip("STG-006", "tiered_storage_awareness", "storage", "compose Kafka services are not available for tiered storage evaluation")
	}

	enabled := 0
	evidence := []string{}
	for _, service := range kafkaServices {
		for _, key := range []string{
			"KAFKA_CFG_REMOTE_LOG_STORAGE_SYSTEM_ENABLE",
			"KAFKA_CFG_REMOTE_STORAGE_ENABLE",
			"KAFKA_CFG_TIERED_STORAGE_ENABLE",
		} {
			value := strings.TrimSpace(service.Environment[key])
			if value == "" {
				continue
			}
			evidence = append(evidence, fmt.Sprintf("service=%s %s=%s", service.ServiceName, key, value))
			if strings.EqualFold(value, "true") || value == "1" {
				enabled++
			}
		}
	}

	if enabled == 0 {
		return rule.NewSkip("STG-006", "tiered_storage_awareness", "storage", "tiered or remote log storage is not enabled in the current compose snapshot")
	}

	result := rule.NewWarn("STG-006", "tiered_storage_awareness", "storage", "tiered or remote log storage is enabled; keep storage diagnostics aware of remote-fetch behavior")
	result.Evidence = evidence
	result.NextActions = []string{"confirm client fetch and retention settings are compatible with remote log reads", "treat storage and latency anomalies with tiered storage context in mind", "review remote log service health before assuming local disk is the only bottleneck"}
	return result
}
