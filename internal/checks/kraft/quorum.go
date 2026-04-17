package kraft

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type QuorumChecker struct{}

func (QuorumChecker) ID() string     { return "KRF-003" }
func (QuorumChecker) Name() string   { return "controller_quorum_majority" }
func (QuorumChecker) Module() string { return "kraft" }

func (QuorumChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Network == nil || len(snap.Network.ControllerChecks) == 0 {
		return rule.NewSkip("KRF-003", "controller_quorum_majority", "kraft", "controller quorum endpoints are not available in the current input mode")
	}

	reachable := 0
	privateControllers := 0
	evidence := make([]string, 0, len(snap.Network.ControllerChecks))
	for _, check := range snap.Network.ControllerChecks {
		if isPrivateEndpoint(check.Address) {
			privateControllers++
		}
		if check.Reachable {
			reachable++
			evidence = append(evidence, fmt.Sprintf("%s reachable", check.Address))
		} else {
			evidence = append(evidence, fmt.Sprintf("%s unreachable: %s", check.Address, check.Error))
		}
	}

	majority := len(snap.Network.ControllerChecks)/2 + 1
	result := rule.NewPass("KRF-003", "controller_quorum_majority", "kraft", "controller quorum has majority")
	result.Evidence = append(evidence, fmt.Sprintf("reachable=%d majority=%d", reachable, majority))
	if reachable < majority {
		if isExternalProbeView(snap) && privateControllers == len(snap.Network.ControllerChecks) {
			result = rule.NewSkip("KRF-003", "controller_quorum_majority", "kraft", "controller quorum cannot be directly verified from the current external probe view")
			result.Evidence = append(evidence, fmt.Sprintf("reachable=%d majority=%d", reachable, majority))
			if snap.Kafka != nil && snap.Kafka.ControllerID != nil {
				result.Evidence = append(result.Evidence, fmt.Sprintf("metadata reports active controller id=%d", *snap.Kafka.ControllerID))
			}
			result.NextActions = []string{"run kdoctor from the Kafka internal network or host", "verify controller listeners from the broker host", "use metadata and broker health as temporary reference"}
			return result
		}
		result = rule.NewCrit("KRF-003", "controller_quorum_majority", "kraft", "controller quorum lost majority")
		result.Evidence = append(evidence, fmt.Sprintf("reachable=%d majority=%d", reachable, majority))
		result.NextActions = []string{"verify controller listeners are reachable", "verify broker-controller processes are healthy", "verify quorum voter configuration"}
	} else if reachable < len(snap.Network.ControllerChecks) {
		result = rule.NewWarn("KRF-003", "controller_quorum_majority", "kraft", "controller quorum still has majority but not all voters are reachable")
		result.Evidence = append(evidence, fmt.Sprintf("reachable=%d majority=%d", reachable, majority))
	}
	return result
}

func isExternalProbeView(snap *snapshot.Bundle) bool {
	if snap == nil || snap.Network == nil {
		return false
	}
	if len(snap.Network.BootstrapChecks) == 0 {
		return false
	}
	for _, check := range snap.Network.BootstrapChecks {
		if !isPrivateEndpoint(check.Address) {
			return true
		}
	}
	return false
}

func isPrivateEndpoint(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	return addr.IsPrivate() || addr.IsLoopback() || addr.IsLinkLocalUnicast()
}
