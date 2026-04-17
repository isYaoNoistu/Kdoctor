package lint

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ClusterIDChecker struct{}

func (ClusterIDChecker) ID() string     { return "CFG-003" }
func (ClusterIDChecker) Name() string   { return "cluster_id_consistency" }
func (ClusterIDChecker) Module() string { return "lint" }

func (ClusterIDChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	services := kafkaServices(getCompose(snap))
	if len(services) == 0 {
		return rule.NewSkip("CFG-003", "cluster_id_consistency", "lint", "compose Kafka services not available")
	}

	var clusterID string
	evidence := []string{}
	for _, service := range services {
		if service.ClusterID == "" {
			result := rule.NewFail("CFG-003", "cluster_id_consistency", "lint", "cluster.id missing in Kafka service")
			result.Evidence = []string{fmt.Sprintf("service=%s", service.ServiceName)}
			return result
		}
		if clusterID == "" {
			clusterID = service.ClusterID
		} else if clusterID != service.ClusterID {
			result := rule.NewFail("CFG-003", "cluster_id_consistency", "lint", "cluster.id is inconsistent across Kafka services")
			result.Evidence = []string{
				fmt.Sprintf("expected=%s got=%s service=%s", clusterID, service.ClusterID, service.ServiceName),
			}
			return result
		}
		evidence = append(evidence, fmt.Sprintf("%s cluster.id=%s", service.ServiceName, service.ClusterID))
	}

	result := rule.NewPass("CFG-003", "cluster_id_consistency", "lint", "cluster.id is consistent across Kafka services")
	result.Evidence = evidence
	return result
}
