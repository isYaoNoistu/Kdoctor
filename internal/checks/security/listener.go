package security

import (
	"context"
	"fmt"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ListenerChecker struct {
	ExecutionView string
	SecurityMode  string
}

func (ListenerChecker) ID() string     { return "SEC-001" }
func (ListenerChecker) Name() string   { return "listener_security_mode" }
func (ListenerChecker) Module() string { return "security" }

func (c ListenerChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	kafkaServices := services(bundle)
	if len(kafkaServices) == 0 {
		return rule.NewSkip("SEC-001", "listener_security_mode", "security", "当前输入模式下没有可用的 compose Kafka 服务，无法评估 listener 安全协议")
	}

	failures := 0
	evidence := make([]string, 0)
	for _, service := range kafkaServices {
		protocols := composeutil.ParseListenerProtocolMap(service.Environment["KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP"])
		if len(protocols) == 0 {
			failures++
			evidence = append(evidence, fmt.Sprintf("服务=%s 缺少 listener.security.protocol.map", service.ServiceName))
			continue
		}

		listenerNames := selectedClientListeners(service, c.ExecutionView)
		if len(listenerNames) == 0 {
			failures++
			evidence = append(evidence, fmt.Sprintf("服务=%s 无法识别当前执行视角对应的客户端 listener", service.ServiceName))
			continue
		}

		for _, listenerName := range listenerNames {
			protocol, ok := protocols[listenerName]
			if !ok {
				failures++
				evidence = append(evidence, fmt.Sprintf("服务=%s listener=%s 未出现在 listener.security.protocol.map 中", service.ServiceName, listenerName))
				continue
			}
			evidence = append(evidence, fmt.Sprintf("服务=%s listener=%s protocol=%s", service.ServiceName, listenerName, protocol))
			if !protocolMatchesMode(c.SecurityMode, protocol) {
				failures++
				evidence = append(evidence, fmt.Sprintf("服务=%s listener=%s 与 profile.security_mode=%s 不匹配", service.ServiceName, listenerName, c.SecurityMode))
			}
		}
	}

	if failures > 0 {
		result := rule.NewFail("SEC-001", "listener_security_mode", "security", "listener 安全协议与当前 profile 执行视角不一致")
		result.Evidence = evidence
		result.NextActions = []string{"核对 profile.security_mode", "检查 KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP", "确认当前执行视角使用的是预期 listener"}
		return result
	}

	result := rule.NewPass("SEC-001", "listener_security_mode", "security", "当前执行视角的 listener 安全协议与 profile 一致")
	result.Evidence = evidence
	return result
}
