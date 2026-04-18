package lint

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ControllerListenerChecker struct{}

func (ControllerListenerChecker) ID() string     { return "CFG-010" }
func (ControllerListenerChecker) Name() string   { return "controller_listener_mapping" }
func (ControllerListenerChecker) Module() string { return "lint" }

func (ControllerListenerChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	services := kafkaServices(getCompose(snap))
	if len(services) == 0 {
		return rule.NewSkip("CFG-010", "controller_listener_mapping", "lint", "compose Kafka services not available")
	}

	failures := 0
	evidence := []string{}
	for _, service := range services {
		listeners, err := parseListeners(service.Listeners)
		if err != nil {
			failures++
			evidence = append(evidence, fmt.Sprintf("service=%s parse_listeners_error=%v", service.ServiceName, err))
			continue
		}
		controllerNames := splitCSV(service.ControllerListenerNames)
		if len(controllerNames) == 0 {
			failures++
			evidence = append(evidence, fmt.Sprintf("service=%s missing controller.listener.names", service.ServiceName))
			continue
		}
		for _, name := range controllerNames {
			if _, ok := listeners[name]; !ok {
				failures++
				evidence = append(evidence, fmt.Sprintf("service=%s controller_listener=%s missing_from_listeners", service.ServiceName, name))
				continue
			}
			evidence = append(evidence, fmt.Sprintf("service=%s controller_listener=%s", service.ServiceName, name))
		}
		voters, err := parseVoters(service.ControllerQuorumVoters)
		if err == nil {
			nodeAddress := voters[service.NodeID]
			if strings.TrimSpace(nodeAddress) != "" {
				matched := false
				for _, name := range controllerNames {
					if listener, ok := listeners[name]; ok && strings.EqualFold(strings.TrimSpace(nodeAddress), fmt.Sprintf("%s:%d", listener.Host, listener.Port)) {
						matched = true
						break
					}
				}
				if !matched {
					failures++
					evidence = append(evidence, fmt.Sprintf("service=%s node_id=%d voter_address=%s does_not_match_controller_listener", service.ServiceName, service.NodeID, nodeAddress))
				}
			}
		}
	}

	result := rule.NewPass("CFG-010", "controller_listener_mapping", "lint", "controller listener names and voter addresses are structurally aligned")
	result.Evidence = evidence
	if failures > 0 {
		result = rule.NewFail("CFG-010", "controller_listener_mapping", "lint", "controller listener names or quorum voter addresses are structurally inconsistent")
		result.Evidence = evidence
		result.NextActions = []string{"align controller.listener.names with the actual listeners block", "make sure each node.id voter address maps to its controller listener endpoint", "avoid reusing broker listeners as controller endpoints by mistake"}
	}
	return result
}
