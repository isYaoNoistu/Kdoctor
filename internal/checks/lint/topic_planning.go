package lint

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type TopicPlanningChecker struct{}

func (TopicPlanningChecker) ID() string     { return "CFG-013" }
func (TopicPlanningChecker) Name() string   { return "topic_defaults_planning" }
func (TopicPlanningChecker) Module() string { return "lint" }

func (TopicPlanningChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	services := kafkaServices(getCompose(snap))
	if len(services) == 0 {
		return rule.NewSkip("CFG-013", "topic_defaults_planning", "lint", "compose Kafka services not available")
	}

	failures := 0
	warnings := 0
	evidence := []string{}
	brokerCount := len(services)
	for _, service := range services {
		evidence = append(evidence,
			fmt.Sprintf("service=%s num_partitions=%d default_rf=%d min_isr=%d", service.ServiceName, service.NumPartitions, service.DefaultReplicationFactor, service.MinISR),
		)
		if service.DefaultReplicationFactor > brokerCount {
			failures++
		}
		if service.NumPartitions > 0 && service.NumPartitions < brokerCount {
			warnings++
		}
		if service.MinISR > service.DefaultReplicationFactor && service.DefaultReplicationFactor > 0 {
			failures++
		}
	}

	result := rule.NewPass("CFG-013", "topic_defaults_planning", "lint", "default topic partition and replication planning looks structurally reasonable")
	result.Evidence = evidence
	switch {
	case failures > 0:
		result = rule.NewFail("CFG-013", "topic_defaults_planning", "lint", "default topic planning contains impossible replication or ISR combinations")
		result.Evidence = evidence
		result.NextActions = []string{"make sure default replication factor does not exceed broker count", "keep min.insync.replicas within the replication factor budget", "review whether default partitions are still sensible for the current topology"}
	case warnings > 0:
		result = rule.NewWarn("CFG-013", "topic_defaults_planning", "lint", "default topic partition planning may be too small for the current broker topology")
		result.Evidence = evidence
		result.NextActions = []string{"review whether default partitions are intentionally smaller than broker count", "adjust topic defaults to avoid accidental hot brokers", "use explicit topic-level planning for critical business topics"}
	}
	return result
}
