package security

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type AuthorizerChecker struct{}

func (AuthorizerChecker) ID() string     { return "SEC-005" }
func (AuthorizerChecker) Name() string   { return "authorizer_config" }
func (AuthorizerChecker) Module() string { return "security" }

func (AuthorizerChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	kafkaServices := services(bundle)
	if len(kafkaServices) == 0 {
		return rule.NewSkip("SEC-005", "authorizer_config", "security", "当前输入模式下没有可用的 compose Kafka 服务，无法评估 Authorizer 配置")
	}

	values := make([]string, 0, len(kafkaServices))
	evidence := make([]string, 0, len(kafkaServices))
	for _, service := range kafkaServices {
		value := strings.TrimSpace(service.Environment["KAFKA_CFG_AUTHORIZER_CLASS_NAME"])
		values = append(values, value)
		if value == "" {
			evidence = append(evidence, fmt.Sprintf("服务=%s 未配置 authorizer.class.name", service.ServiceName))
			continue
		}
		evidence = append(evidence, fmt.Sprintf("服务=%s authorizer=%s", service.ServiceName, value))
		if !strings.Contains(strings.ToLower(value), "standardauthorizer") {
			result := rule.NewFail("SEC-005", "authorizer_config", "security", "检测到非 StandardAuthorizer 的 Authorizer 配置")
			result.Evidence = evidence
			result.NextActions = []string{"在 KRaft 场景优先使用 StandardAuthorizer", "检查 authorizer.class.name 是否为预期配置", "确认 ACL 策略是否已经迁移完成"}
			return result
		}
	}

	if inconsistent(values) {
		result := rule.NewFail("SEC-005", "authorizer_config", "security", "Kafka 服务之间的 Authorizer 配置不一致")
		result.Evidence = evidence
		result.NextActions = []string{"统一各 broker 的 authorizer.class.name", "确认所有节点使用相同 ACL 策略", "重启前先核对 controller 与 broker 配置一致性"}
		return result
	}

	if allEmpty(values) {
		result := rule.NewWarn("SEC-005", "authorizer_config", "security", "当前未配置 Authorizer，ACL 拒绝类问题将无法被治理")
		result.Evidence = evidence
		result.NextActions = []string{"如果集群需要 ACL，请启用 StandardAuthorizer", "确认当前环境是否允许无 ACL 运行", "把授权策略纳入后续 V2 安全域基线"}
		return result
	}

	result := rule.NewPass("SEC-005", "authorizer_config", "security", "Authorizer 配置一致，且已使用 StandardAuthorizer")
	result.Evidence = evidence
	return result
}

func inconsistent(values []string) bool {
	normalized := ""
	for _, value := range values {
		value = strings.TrimSpace(value)
		if normalized == "" {
			normalized = value
			continue
		}
		if normalized != value {
			return true
		}
	}
	return false
}

func allEmpty(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
