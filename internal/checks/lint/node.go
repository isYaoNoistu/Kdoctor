package lint

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type NodeIDChecker struct{}

func (NodeIDChecker) ID() string     { return "CFG-002" }
func (NodeIDChecker) Name() string   { return "node_id_uniqueness" }
func (NodeIDChecker) Module() string { return "lint" }

func (NodeIDChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	services := kafkaServices(getCompose(snap))
	if len(services) == 0 {
		return rule.NewSkip("CFG-002", "node_id_uniqueness", "lint", "compose Kafka services not available")
	}

	seen := map[int]string{}
	evidence := []string{}
	for _, service := range services {
		if service.NodeIDRaw == "" {
			result := rule.NewFail("CFG-002", "node_id_uniqueness", "lint", "node.id missing in Kafka service")
			result.Evidence = []string{fmt.Sprintf("service=%s", service.ServiceName)}
			return result
		}
		if previous, ok := seen[service.NodeID]; ok {
			result := rule.NewFail("CFG-002", "node_id_uniqueness", "lint", "duplicate node.id detected")
			result.Evidence = []string{
				fmt.Sprintf("node.id=%d used by %s and %s", service.NodeID, previous, service.ServiceName),
			}
			return result
		}
		seen[service.NodeID] = service.ServiceName
		evidence = append(evidence, fmt.Sprintf("%s node.id=%d", service.ServiceName, service.NodeID))
	}

	result := rule.NewPass("CFG-002", "node_id_uniqueness", "lint", "node.id values are unique")
	result.Evidence = evidence
	return result
}
