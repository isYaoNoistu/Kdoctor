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

type BootstrapLBChecker struct{}

func (BootstrapLBChecker) ID() string     { return "NET-007" }
func (BootstrapLBChecker) Name() string   { return "bootstrap_only_load_balancer" }
func (BootstrapLBChecker) Module() string { return "network" }

func (BootstrapLBChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Network == nil {
		return rule.NewSkip("NET-007", "bootstrap_only_load_balancer", "network", "当前没有可用的网络快照，无法评估 bootstrap-only LB 场景")
	}

	bootstrapHosts := reachableHosts(snap.Network.BootstrapChecks)
	if len(bootstrapHosts) != 1 {
		return rule.NewSkip("NET-007", "bootstrap_only_load_balancer", "network", "bootstrap 地址不呈现单入口形态，暂不按 bootstrap-only LB 场景判断")
	}
	if len(snap.Network.MetadataChecks) == 0 {
		return rule.NewSkip("NET-007", "bootstrap_only_load_balancer", "network", "尚未执行 metadata 返回端点探测")
	}

	unreachableMetadata := 0
	distinctMetadataHosts := hostSet(snap.Network.MetadataChecks)
	evidence := []string{fmt.Sprintf("bootstrap_host=%s", bootstrapHosts[0])}
	for _, check := range snap.Network.MetadataChecks {
		if !check.Reachable {
			unreachableMetadata++
			evidence = append(evidence, fmt.Sprintf("metadata 端点=%s 不可达", check.Address))
		}
	}
	for _, host := range distinctMetadataHosts {
		evidence = append(evidence, fmt.Sprintf("metadata_host=%s", host))
	}

	if unreachableMetadata == 0 || len(distinctMetadataHosts) == 0 {
		return rule.NewPass("NET-007", "bootstrap_only_load_balancer", "network", "未发现仅代理 bootstrap 的负载均衡迹象")
	}
	if len(distinctMetadataHosts) == 1 && distinctMetadataHosts[0] == bootstrapHosts[0] {
		return rule.NewPass("NET-007", "bootstrap_only_load_balancer", "network", "bootstrap 与 metadata 主机路径一致，未见 LB 入口错配")
	}

	result := rule.NewWarn("NET-007", "bootstrap_only_load_balancer", "network", "疑似存在只代理 bootstrap 的 LB，metadata 后续流量已经绕到 broker 直连地址")
	result.Evidence = evidence
	result.NextActions = []string{"确认负载均衡是否只覆盖 bootstrap 入口", "评估客户端是否应直连所有 broker 地址", "核对 advertised.listeners 应返回 LB 地址还是 broker 直连地址"}
	return result
}

func reachableHosts(checks []snapshot.EndpointCheck) []string {
	hosts := make([]string, 0)
	seen := map[string]struct{}{}
	for _, check := range checks {
		if !check.Reachable {
			continue
		}
		host, _, err := net.SplitHostPort(check.Address)
		if err != nil {
			host = check.Address
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)
	return hosts
}

func hostSet(checks []snapshot.EndpointCheck) []string {
	hosts := make([]string, 0)
	seen := map[string]struct{}{}
	for _, check := range checks {
		host, _, err := net.SplitHostPort(check.Address)
		if err != nil {
			host = check.Address
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)
	return hosts
}
