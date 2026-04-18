package host

import (
	"context"
	"fmt"
	"net"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ListenerDriftChecker struct{}

func (ListenerDriftChecker) ID() string     { return "HOST-010" }
func (ListenerDriftChecker) Name() string   { return "listener_port_drift" }
func (ListenerDriftChecker) Module() string { return "host" }

func (ListenerDriftChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Host == nil || !bundle.Host.Collected || len(bundle.Host.ObservedListenPorts) == 0 {
		return rule.NewSkip("HOST-010", "listener_port_drift", "host", "actual listening port evidence is not available in the current input mode")
	}
	if len(bundle.Host.PortChecks) == 0 {
		return rule.NewSkip("HOST-010", "listener_port_drift", "host", "expected listener ports are not available for drift comparison")
	}

	actual := map[int]struct{}{}
	for _, port := range bundle.Host.ObservedListenPorts {
		actual[port] = struct{}{}
	}

	missing := 0
	evidence := []string{}
	for _, check := range bundle.Host.PortChecks {
		port := portFromAddress(check.Address)
		if port <= 0 {
			continue
		}
		_, listening := actual[port]
		evidence = append(evidence, fmt.Sprintf("expected_port=%d listening=%t reachable=%t", port, listening, check.Reachable))
		if !listening {
			missing++
		}
	}

	result := rule.NewPass("HOST-010", "listener_port_drift", "host", "expected listener ports are present in the current host listening table")
	result.Evidence = evidence
	if missing > 0 {
		result = rule.NewFail("HOST-010", "listener_port_drift", "host", "some expected Kafka listener ports are missing from the host listening table")
		result.Evidence = evidence
		result.NextActions = []string{"compare ss or netstat output with Kafka listener configuration", "check whether the broker process bound a different port or address than expected", "confirm Docker host-network and listener exposure match the current runtime state"}
	}
	return result
}

func portFromAddress(address string) int {
	_, portText, err := net.SplitHostPort(address)
	if err != nil {
		return 0
	}
	var port int
	fmt.Sscanf(portText, "%d", &port)
	return port
}
