package kraft

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ConfigChecker struct{}

func (ConfigChecker) ID() string     { return "KRF-001" }
func (ConfigChecker) Name() string   { return "controller_quorum_config_consistency" }
func (ConfigChecker) Module() string { return "kraft" }

func (ConfigChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	services := composeutil.KafkaServices(getCompose(snap))
	if len(services) == 0 {
		if snap == nil || snap.Network == nil || len(snap.Network.ControllerChecks) == 0 {
			return rule.NewSkip("KRF-001", "controller_quorum_config_consistency", "kraft", "controller quorum configuration is not available in the current input mode")
		}
		result := rule.NewPass("KRF-001", "controller_quorum_config_consistency", "kraft", "controller quorum endpoints were provided explicitly")
		for _, check := range snap.Network.ControllerChecks {
			result.Evidence = append(result.Evidence, fmt.Sprintf("controller endpoint=%s", check.Address))
		}
		return result
	}

	var baseline string
	var baselineVoters map[int]string
	controllerNodeIDs := map[int]struct{}{}
	evidence := []string{}
	for _, service := range services {
		roles := composeutil.ParseCSV(service.Environment["KAFKA_CFG_PROCESS_ROLES"])
		if contains(roles, "controller") {
			nodeID, err := strconv.Atoi(strings.TrimSpace(service.Environment["KAFKA_CFG_NODE_ID"]))
			if err != nil {
				return rule.NewFail("KRF-001", "controller_quorum_config_consistency", "kraft", "controller node.id is missing or invalid in compose")
			}
			controllerNodeIDs[nodeID] = struct{}{}
		}

		votersRaw := strings.TrimSpace(service.Environment["KAFKA_CFG_CONTROLLER_QUORUM_VOTERS"])
		if votersRaw == "" {
			result := rule.NewFail("KRF-001", "controller_quorum_config_consistency", "kraft", "controller.quorum.voters is missing in a Kafka service")
			result.Evidence = []string{fmt.Sprintf("service=%s", service.ServiceName)}
			return result
		}

		voters, err := composeutil.ParseVoters(votersRaw)
		if err != nil {
			result := rule.NewFail("KRF-001", "controller_quorum_config_consistency", "kraft", "controller.quorum.voters format is invalid")
			result.Evidence = []string{fmt.Sprintf("service=%s value=%s", service.ServiceName, votersRaw)}
			return result
		}
		if baseline == "" {
			baseline = votersRaw
			baselineVoters = voters
		} else if baseline != votersRaw {
			result := rule.NewFail("KRF-001", "controller_quorum_config_consistency", "kraft", "controller.quorum.voters differs across Kafka services")
			result.Evidence = []string{
				fmt.Sprintf("baseline=%s", baseline),
				fmt.Sprintf("service=%s voters=%s", service.ServiceName, votersRaw),
			}
			return result
		}
		evidence = append(evidence, fmt.Sprintf("%s voters=%s", service.ServiceName, votersRaw))
	}

	missingControllers := []int{}
	for nodeID := range controllerNodeIDs {
		if _, ok := baselineVoters[nodeID]; !ok {
			missingControllers = append(missingControllers, nodeID)
		}
	}
	sort.Ints(missingControllers)
	if len(missingControllers) > 0 {
		result := rule.NewFail("KRF-001", "controller_quorum_config_consistency", "kraft", "some controller node.id values are missing from controller.quorum.voters")
		for _, nodeID := range missingControllers {
			result.Evidence = append(result.Evidence, fmt.Sprintf("missing controller node.id=%d", nodeID))
		}
		return result
	}

	if snap != nil && snap.Network != nil && len(snap.Network.ControllerChecks) > 0 {
		expected := sortedVoterAddresses(baselineVoters)
		actual := sortedControllerChecks(snap.Network.ControllerChecks)
		if strings.Join(expected, ",") != strings.Join(actual, ",") {
			result := rule.NewWarn("KRF-001", "controller_quorum_config_consistency", "kraft", "explicit controller endpoints differ from compose quorum voters")
			result.Evidence = append(evidence,
				fmt.Sprintf("compose voters=%v", expected),
				fmt.Sprintf("runtime endpoints=%v", actual),
			)
			result.NextActions = []string{"align profile controller_endpoints with compose quorum voters", "verify the current execution view uses the intended controller listener addresses"}
			return result
		}
	}

	result := rule.NewPass("KRF-001", "controller_quorum_config_consistency", "kraft", "controller quorum configuration is consistent")
	result.Evidence = evidence
	return result
}

func getCompose(snap *snapshot.Bundle) *snapshot.ComposeSnapshot {
	if snap == nil {
		return nil
	}
	return snap.Compose
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func sortedVoterAddresses(voters map[int]string) []string {
	out := make([]string, 0, len(voters))
	for _, address := range voters {
		out = append(out, address)
	}
	sort.Strings(out)
	return out
}

func sortedControllerChecks(checks []snapshot.EndpointCheck) []string {
	out := make([]string, 0, len(checks))
	for _, check := range checks {
		out = append(out, check.Address)
	}
	sort.Strings(out)
	return out
}
