package network

import (
	"context"
	"fmt"
	"net"
	"sort"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type DNSChecker struct{}

func (DNSChecker) ID() string     { return "NET-004" }
func (DNSChecker) Name() string   { return "dns_resolution" }
func (DNSChecker) Module() string { return "network" }

func (DNSChecker) Run(ctx context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Network == nil {
		return rule.NewError("NET-004", "dns_resolution", "network", "DNS resolution cannot be evaluated", "network snapshot missing")
	}

	hosts := collectHosts(snap.Network)
	if len(hosts) == 0 {
		return rule.NewSkip("NET-004", "dns_resolution", "network", "no endpoint hosts are available for DNS resolution")
	}

	allLiteral := true
	evidence := []string{}
	failures := 0
	for _, host := range hosts {
		if net.ParseIP(host) != nil {
			evidence = append(evidence, fmt.Sprintf("%s literal-ip", host))
			continue
		}
		allLiteral = false
		addrs, err := net.DefaultResolver.LookupHost(ctx, host)
		if err != nil {
			failures++
			evidence = append(evidence, fmt.Sprintf("%s lookup failed: %v", host, err))
			continue
		}
		sort.Strings(addrs)
		evidence = append(evidence, fmt.Sprintf("%s -> %v", host, addrs))
	}

	if allLiteral {
		result := rule.NewSkip("NET-004", "dns_resolution", "network", "all configured endpoints use literal IP addresses")
		result.Evidence = evidence
		return result
	}

	result := rule.NewPass("NET-004", "dns_resolution", "network", "all configured hostnames resolve successfully")
	result.Evidence = evidence
	if failures > 0 {
		result = rule.NewFail("NET-004", "dns_resolution", "network", "some configured hostnames failed DNS resolution")
		result.Evidence = evidence
		result.NextActions = []string{"verify DNS records for the failing hostnames", "compare bootstrap and advertised listener hostnames", "use literal IP addresses temporarily if DNS is unstable"}
	}
	return result
}

func collectHosts(network *snapshot.NetworkSnapshot) []string {
	if network == nil {
		return nil
	}
	seen := map[string]struct{}{}
	out := []string{}
	appendHost := func(address string) {
		host, _, err := net.SplitHostPort(address)
		if err != nil || host == "" {
			host = address
		}
		if _, ok := seen[host]; ok {
			return
		}
		seen[host] = struct{}{}
		out = append(out, host)
	}
	for _, check := range network.BootstrapChecks {
		appendHost(check.Address)
	}
	for _, check := range network.ControllerChecks {
		appendHost(check.Address)
	}
	for _, check := range network.MetadataChecks {
		appendHost(check.Address)
	}
	sort.Strings(out)
	return out
}
