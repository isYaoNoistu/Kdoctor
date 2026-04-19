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
		return rule.NewSkip("HOST-010", "listener_port_drift", "host", "当前输入模式下没有可用的实际监听端口证据")
	}
	if len(bundle.Host.PortChecks) == 0 {
		return rule.NewSkip("HOST-010", "listener_port_drift", "host", "当前没有可用于对比的期望 listener 端口")
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

	result := rule.NewPass("HOST-010", "listener_port_drift", "host", "期望的 listener 端口都出现在当前宿主机监听表中")
	result.Evidence = evidence
	if missing > 0 {
		result = rule.NewFail("HOST-010", "listener_port_drift", "host", "部分期望的 Kafka listener 端口没有出现在宿主机监听表中")
		result.Evidence = evidence
		result.NextActions = []string{"对比 ss 或 netstat 输出与 Kafka listener 配置", "检查 broker 进程是否绑定到了与预期不同的端口或地址", "确认 Docker host-network 和 listener 暴露与当前运行状态一致"}
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
