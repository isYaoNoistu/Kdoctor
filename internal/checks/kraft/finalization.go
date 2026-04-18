package kraft

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type FinalizationChecker struct{}

func (FinalizationChecker) ID() string     { return "KRF-008" }
func (FinalizationChecker) Name() string   { return "kraft_finalization_state" }
func (FinalizationChecker) Module() string { return "kraft" }

func (FinalizationChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Compose == nil {
		return rule.NewSkip("KRF-008", "kraft_finalization_state", "kraft", "compose snapshot is not available for KRaft finalization evaluation")
	}

	services := composeutil.KafkaServices(bundle.Compose)
	if len(services) == 0 {
		return rule.NewSkip("KRF-008", "kraft_finalization_state", "kraft", "compose Kafka services are not available")
	}

	kraftVersions := map[string][]string{}
	metadataVersions := map[string][]string{}
	bootstrapServices := []string{}
	for _, service := range services {
		if value := strings.TrimSpace(service.Environment["KAFKA_KRAFT_VERSION"]); value != "" {
			kraftVersions[value] = append(kraftVersions[value], service.ServiceName)
		}
		if value := strings.TrimSpace(service.Environment["KAFKA_CFG_METADATA_VERSION"]); value != "" {
			metadataVersions[value] = append(metadataVersions[value], service.ServiceName)
		}
		if strings.TrimSpace(service.Environment["KAFKA_CFG_CONTROLLER_QUORUM_BOOTSTRAP_SERVERS"]) != "" {
			bootstrapServices = append(bootstrapServices, service.ServiceName)
		}
	}

	evidence := []string{}
	appendVersionEvidence := func(prefix string, groups map[string][]string) {
		keys := make([]string, 0, len(groups))
		for key := range groups {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			evidence = append(evidence, fmt.Sprintf("%s=%s services=%v", prefix, key, groups[key]))
		}
	}
	appendVersionEvidence("kraft_version", kraftVersions)
	appendVersionEvidence("metadata_version", metadataVersions)
	if len(bootstrapServices) > 0 {
		sort.Strings(bootstrapServices)
		evidence = append(evidence, fmt.Sprintf("controller_quorum_bootstrap_servers=%v", bootstrapServices))
	}

	result := rule.NewPass("KRF-008", "kraft_finalization_state", "kraft", "KRaft finalization related configuration does not show an obvious unfinished migration state")
	result.Evidence = evidence
	switch {
	case len(kraftVersions) > 1 || len(metadataVersions) > 1:
		result = rule.NewWarn("KRF-008", "kraft_finalization_state", "kraft", "KRaft or metadata feature versions are inconsistent across brokers and may indicate an unfinished finalization or upgrade")
		result.Evidence = evidence
		result.NextActions = []string{"verify whether the upgrade window is still in progress", "align KRaft and metadata feature versions once the rollout is complete", "remove stale compatibility settings after confirming controller convergence"}
	case len(bootstrapServices) > 0:
		result = rule.NewWarn("KRF-008", "kraft_finalization_state", "kraft", "controller.quorum.bootstrap.servers is still configured; confirm whether the KRaft migration and finalization process has fully closed")
		result.Evidence = evidence
		result.NextActions = []string{"confirm whether this cluster intentionally keeps bootstrap server compatibility settings", "review KRaft migration status and finalization notes", "clean up transitional controller quorum settings if the migration already ended"}
	}
	return result
}
