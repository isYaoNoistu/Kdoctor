package lint

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type AdvertisedViewChecker struct {
	ExecutionView string
}

func (AdvertisedViewChecker) ID() string     { return "CFG-009" }
func (AdvertisedViewChecker) Name() string   { return "advertised_listener_view" }
func (AdvertisedViewChecker) Module() string { return "lint" }

func (c AdvertisedViewChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	services := kafkaServices(getCompose(snap))
	if len(services) == 0 {
		return rule.NewSkip("CFG-009", "advertised_listener_view", "lint", "compose Kafka services not available")
	}
	if !strings.EqualFold(strings.TrimSpace(c.ExecutionView), "external") {
		return rule.NewSkip("CFG-009", "advertised_listener_view", "lint", "current profile execution_view is not external")
	}

	privateOnly := 0
	evidence := []string{}
	for _, service := range services {
		advertised, err := parseListeners(service.AdvertisedListeners)
		if err != nil {
			continue
		}
		external := advertised["EXTERNAL"]
		host := strings.TrimSpace(external.Host)
		evidence = append(evidence, fmt.Sprintf("service=%s advertised_external=%s", service.ServiceName, external.Raw))
		if isPrivateHost(host) {
			privateOnly++
		}
	}

	result := rule.NewPass("CFG-009", "advertised_listener_view", "lint", "external execution view is compatible with the configured advertised listeners")
	result.Evidence = evidence
	if privateOnly > 0 {
		result = rule.NewFail("CFG-009", "advertised_listener_view", "lint", "some EXTERNAL advertised listeners still point to private addresses")
		result.Evidence = evidence
		result.NextActions = []string{"publish routable EXTERNAL listener addresses for public or cross-network clients", "avoid mixing private addresses into the external listener set", "verify advertised.listeners from the same network where clients will run"}
	}
	return result
}

func isPrivateHost(host string) bool {
	addr, err := netip.ParseAddr(strings.TrimSpace(host))
	if err != nil {
		return false
	}
	return addr.IsPrivate() || addr.IsLoopback() || addr.IsLinkLocalUnicast()
}
