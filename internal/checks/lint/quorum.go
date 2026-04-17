package lint

import (
	"context"
	"fmt"
	"sort"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type QuorumVotersChecker struct{}

func (QuorumVotersChecker) ID() string     { return "CFG-005" }
func (QuorumVotersChecker) Name() string   { return "controller_quorum_voters_consistency" }
func (QuorumVotersChecker) Module() string { return "lint" }

func (QuorumVotersChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	services := kafkaServices(getCompose(snap))
	if len(services) == 0 {
		return rule.NewSkip("CFG-005", "controller_quorum_voters_consistency", "lint", "compose Kafka services not available")
	}

	var baseline string
	var baselineVoters map[int]string
	nodeIDs := map[int]struct{}{}
	evidence := []string{}
	for _, service := range services {
		if service.ControllerQuorumVoters == "" {
			result := rule.NewFail("CFG-005", "controller_quorum_voters_consistency", "lint", "controller.quorum.voters missing in Kafka service")
			result.Evidence = []string{fmt.Sprintf("service=%s", service.ServiceName)}
			return result
		}
		voters, err := parseVoters(service.ControllerQuorumVoters)
		if err != nil {
			result := rule.NewFail("CFG-005", "controller_quorum_voters_consistency", "lint", "controller.quorum.voters format is invalid")
			result.Evidence = []string{fmt.Sprintf("service=%s error=%v", service.ServiceName, err)}
			return result
		}
		if baseline == "" {
			baseline = service.ControllerQuorumVoters
			baselineVoters = voters
		} else if baseline != service.ControllerQuorumVoters {
			result := rule.NewFail("CFG-005", "controller_quorum_voters_consistency", "lint", "controller.quorum.voters is inconsistent across Kafka services")
			result.Evidence = []string{
				fmt.Sprintf("baseline=%s", baseline),
				fmt.Sprintf("service=%s voters=%s", service.ServiceName, service.ControllerQuorumVoters),
			}
			return result
		}
		nodeIDs[service.NodeID] = struct{}{}
		evidence = append(evidence, fmt.Sprintf("%s voters=%s", service.ServiceName, service.ControllerQuorumVoters))
	}

	var missing []int
	for nodeID := range nodeIDs {
		if _, ok := baselineVoters[nodeID]; !ok {
			missing = append(missing, nodeID)
		}
	}
	sort.Ints(missing)
	if len(missing) > 0 {
		result := rule.NewFail("CFG-005", "controller_quorum_voters_consistency", "lint", "some node.id values are not represented in controller.quorum.voters")
		for _, nodeID := range missing {
			result.Evidence = append(result.Evidence, fmt.Sprintf("missing node.id=%d", nodeID))
		}
		return result
	}

	result := rule.NewPass("CFG-005", "controller_quorum_voters_consistency", "lint", "controller.quorum.voters is consistent and matches node.id values")
	result.Evidence = evidence
	return result
}
