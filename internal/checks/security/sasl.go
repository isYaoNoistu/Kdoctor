package security

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type SASLChecker struct {
	ExecutionView string
	SecurityMode  string
	SASLMechanism string
}

func (SASLChecker) ID() string     { return "SEC-002" }
func (SASLChecker) Name() string   { return "sasl_mechanism_consistency" }
func (SASLChecker) Module() string { return "security" }

func (c SASLChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	kafkaServices := services(bundle)
	if len(kafkaServices) == 0 {
		return rule.NewSkip("SEC-002", "sasl_mechanism_consistency", "security", "当前输入模式下没有可用的 compose Kafka 服务，无法评估 SASL 机制")
	}

	requiredMechanism := strings.ToUpper(strings.TrimSpace(c.SASLMechanism))
	needsSASL := strings.HasPrefix(normalizeSecurityMode(c.SecurityMode), "sasl")
	failures := 0
	evidence := make([]string, 0)

	for _, service := range kafkaServices {
		listenerNames := selectedClientListeners(service, c.ExecutionView)
		protocols := composeutil.ParseListenerProtocolMap(service.Environment["KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP"])
		serviceNeedsSASL := needsSASL
		for _, listenerName := range listenerNames {
			if protocolNeedsSASL(protocols[listenerName]) {
				serviceNeedsSASL = true
			}
		}
		if !serviceNeedsSASL {
			continue
		}

		mechanisms := splitUpperCSV(service.Environment["KAFKA_CFG_SASL_ENABLED_MECHANISMS"])
		if len(mechanisms) == 0 {
			failures++
			evidence = append(evidence, fmt.Sprintf("服务=%s 需要 SASL，但未配置 KAFKA_CFG_SASL_ENABLED_MECHANISMS", service.ServiceName))
			continue
		}

		evidence = append(evidence, fmt.Sprintf("服务=%s enabled_mechanisms=%s", service.ServiceName, strings.Join(mechanisms, ",")))
		if requiredMechanism != "" && !contains(mechanisms, requiredMechanism) {
			failures++
			evidence = append(evidence, fmt.Sprintf("服务=%s 缺少 profile 指定的 SASL 机制=%s", service.ServiceName, requiredMechanism))
		}
	}

	if !needsSASL && failures == 0 && len(evidence) == 0 {
		return rule.NewSkip("SEC-002", "sasl_mechanism_consistency", "security", "当前 profile/compose 未启用 SASL listener")
	}

	if failures > 0 {
		result := rule.NewFail("SEC-002", "sasl_mechanism_consistency", "security", "SASL 机制配置与当前 listener 安全模式不一致")
		result.Evidence = evidence
		result.NextActions = []string{"检查 KAFKA_CFG_SASL_ENABLED_MECHANISMS", "确认 profile.sasl_mechanism 与 broker 配置一致", "核对当前执行视角是否真的应走 SASL listener"}
		return result
	}

	summary := "SASL 机制配置与当前 listener 安全模式一致"
	if requiredMechanism != "" {
		summary = fmt.Sprintf("SASL 机制配置已覆盖 profile 指定的 %s", requiredMechanism)
	}
	result := rule.NewPass("SEC-002", "sasl_mechanism_consistency", "security", summary)
	result.Evidence = evidence
	return result
}
