package network

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type DNSDriftChecker struct{}

func (DNSDriftChecker) ID() string     { return "NET-008" }
func (DNSDriftChecker) Name() string   { return "dns_drift" }
func (DNSDriftChecker) Module() string { return "network" }

func (DNSDriftChecker) Run(ctx context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Network == nil {
		return rule.NewSkip("NET-008", "dns_drift", "network", "network snapshot is not available")
	}

	hosts := collectHosts(bundle.Network)
	targetHosts := make([]string, 0)
	for _, host := range hosts {
		if net.ParseIP(host) == nil {
			targetHosts = append(targetHosts, host)
		}
	}
	if len(targetHosts) == 0 {
		return rule.NewSkip("NET-008", "dns_drift", "network", "no hostname-based endpoints are available for DNS drift analysis")
	}

	metadataHostSet := resolvedMetadataHostSet(ctx, bundle)
	multiValue := 0
	drift := 0
	evidence := []string{}
	for _, host := range targetHosts {
		addrs, err := net.DefaultResolver.LookupHost(ctx, host)
		if err != nil {
			evidence = append(evidence, fmt.Sprintf("host=%s lookup_error=%v", host, err))
			continue
		}
		sort.Strings(addrs)
		if len(addrs) > 1 {
			multiValue++
		}
		evidence = append(evidence, fmt.Sprintf("host=%s resolved=%v", host, addrs))
		if len(metadataHostSet) == 0 {
			continue
		}
		if !hasIntersection(addrs, metadataHostSet) {
			drift++
			evidence = append(evidence, fmt.Sprintf("host=%s has no overlap with metadata_hosts=%v", host, sortedKeys(metadataHostSet)))
		}
	}

	result := rule.NewPass("NET-008", "dns_drift", "network", "hostname resolution is broadly consistent with the current Kafka route view")
	result.Evidence = evidence
	switch {
	case drift > 0:
		result = rule.NewWarn("NET-008", "dns_drift", "network", "DNS resolution differs from the current metadata route view and may indicate stale records or split routing")
		result.Evidence = evidence
		result.NextActions = []string{"compare DNS A records with advertised.listeners and metadata returned addresses", "check whether load balancer, NAT, or DNS cache still points to an old broker address", "prefer the route set that matches the current execution view"}
	case multiValue > 0:
		result = rule.NewWarn("NET-008", "dns_drift", "network", "some Kafka hostnames resolve to multiple addresses; verify that all returned routes are intentional")
		result.Evidence = evidence
		result.NextActions = []string{"confirm multi-value DNS is expected for the current deployment", "check whether clients and brokers use the same hostname resolution path", "verify old IP addresses have been removed after topology changes"}
	}
	return result
}

func resolvedMetadataHostSet(ctx context.Context, bundle *snapshot.Bundle) map[string]struct{} {
	out := map[string]struct{}{}
	if bundle == nil || bundle.Network == nil {
		return out
	}
	for _, check := range bundle.Network.MetadataChecks {
		host, _, err := net.SplitHostPort(check.Address)
		if err != nil {
			host = check.Address
		}
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}
		if ip := net.ParseIP(host); ip != nil {
			out[ip.String()] = struct{}{}
			continue
		}
		addrs, err := net.DefaultResolver.LookupHost(ctx, host)
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			out[addr] = struct{}{}
		}
	}
	return out
}

func hasIntersection(values []string, set map[string]struct{}) bool {
	for _, value := range values {
		if _, ok := set[value]; ok {
			return true
		}
	}
	return false
}

func sortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
