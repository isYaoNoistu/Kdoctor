package network

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type RouteMismatchChecker struct{}

func (RouteMismatchChecker) ID() string     { return "NET-005" }
func (RouteMismatchChecker) Name() string   { return "metadata_route_mismatch" }
func (RouteMismatchChecker) Module() string { return "network" }

func (RouteMismatchChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Network == nil {
		return rule.NewSkip("NET-005", "metadata_route_mismatch", "network", "当前没有可用的网络快照，无法评估 metadata 路由问题")
	}

	bootstrapOK := 0
	for _, check := range snap.Network.BootstrapChecks {
		if check.Reachable {
			bootstrapOK++
		}
	}
	if bootstrapOK == 0 {
		return rule.NewSkip("NET-005", "metadata_route_mismatch", "network", "bootstrap 本身不可达，暂不评估 metadata 路由错配")
	}
	if len(snap.Network.MetadataChecks) == 0 {
		return rule.NewSkip("NET-005", "metadata_route_mismatch", "network", "尚未执行 metadata 返回端点探测")
	}

	unreachable := 0
	evidence := []string{fmt.Sprintf("bootstrap_reachable=%d", bootstrapOK)}
	for _, check := range snap.Network.MetadataChecks {
		if check.Reachable {
			evidence = append(evidence, fmt.Sprintf("metadata=%s reachable", check.Address))
			continue
		}
		unreachable++
		evidence = append(evidence, fmt.Sprintf("metadata=%s unreachable: %s", check.Address, check.Error))
	}

	result := rule.NewPass("NET-005", "metadata_route_mismatch", "network", "bootstrap 与 metadata 返回地址的网络路径一致")
	result.Evidence = evidence
	if unreachable == len(snap.Network.MetadataChecks) {
		result = rule.NewFail("NET-005", "metadata_route_mismatch", "network", "bootstrap 可达，但 metadata 返回地址整体不可达，更像是 returned address 路由错配")
		result.Evidence = evidence
		result.NextActions = []string{"优先核对 advertised.listeners", "检查客户端所在网络到 broker 实际地址的路由", "确认负载均衡是否只代理 bootstrap 入口"}
		return result
	}
	if unreachable > 0 {
		result = rule.NewWarn("NET-005", "metadata_route_mismatch", "network", "bootstrap 可达，但部分 metadata 返回地址不可达，存在路由或 listeners 错配")
		result.Evidence = evidence
		result.NextActions = []string{"核对不可达 broker 的 advertised.listeners", "检查端口暴露、NAT 与安全组", "确认当前执行视角是否与 broker 返回地址匹配"}
	}
	return result
}
