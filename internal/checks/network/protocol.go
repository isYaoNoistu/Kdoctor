package network

import (
	"context"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ProtocolMismatchChecker struct{}

func (ProtocolMismatchChecker) ID() string     { return "NET-009" }
func (ProtocolMismatchChecker) Name() string   { return "socket_open_but_kafka_handshake_failed" }
func (ProtocolMismatchChecker) Module() string { return "network" }

func (ProtocolMismatchChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Network == nil {
		return rule.NewSkip("NET-009", "socket_open_but_kafka_handshake_failed", "network", "当前没有可用的网络快照，无法评估协议错配")
	}

	reachableBootstrap := 0
	for _, check := range snap.Network.BootstrapChecks {
		if check.Reachable {
			reachableBootstrap++
		}
	}
	if reachableBootstrap == 0 {
		return rule.NewSkip("NET-009", "socket_open_but_kafka_handshake_failed", "network", "bootstrap 本身不可达，暂不评估协议错配")
	}
	if snap.Kafka != nil {
		return rule.NewPass("NET-009", "socket_open_but_kafka_handshake_failed", "network", "TCP 连通后 Kafka metadata 也可获取，未见明显协议错配")
	}

	result := rule.NewFail("NET-009", "socket_open_but_kafka_handshake_failed", "network", "TCP 已打开，但 Kafka metadata 获取失败，存在协议、安全或握手层错配")
	result.Evidence = []string{"bootstrap TCP 可达，但未形成 Kafka metadata 快照"}
	result.NextActions = []string{"检查 listener.security.protocol.map 与客户端安全模式", "确认端口后面确实是 Kafka 协议而不是仅端口可连", "排查 SASL/SSL 握手、反向代理或四层转发错配"}
	return result
}
