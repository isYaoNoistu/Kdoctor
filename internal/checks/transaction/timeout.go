package transaction

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type TimeoutChecker struct {
	TransactionTimeoutMs int
}

func (TimeoutChecker) ID() string     { return "TXN-003" }
func (TimeoutChecker) Name() string   { return "transaction_timeout_limit" }
func (TimeoutChecker) Module() string { return "transaction" }

func (c TimeoutChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if c.TransactionTimeoutMs == 0 {
		return rule.NewSkip("TXN-003", "transaction_timeout_limit", "transaction", "当前未配置 transaction.timeout.ms，暂不评估事务超时限制")
	}

	evidence := []string{fmt.Sprintf("transaction_timeout_ms=%d", c.TransactionTimeoutMs)}
	brokerLimit := 0
	if bundle != nil && bundle.Compose != nil {
		for _, service := range composeutil.KafkaServices(bundle.Compose) {
			if raw := strings.TrimSpace(service.Environment["KAFKA_CFG_TRANSACTION_MAX_TIMEOUT_MS"]); raw != "" {
				if value, err := strconv.Atoi(raw); err == nil {
					evidence = append(evidence, fmt.Sprintf("service=%s transaction_max_timeout_ms=%d", service.ServiceName, value))
					if brokerLimit == 0 || value < brokerLimit {
						brokerLimit = value
					}
				}
			}
		}
	}
	if brokerLimit > 0 && c.TransactionTimeoutMs > brokerLimit {
		result := rule.NewFail("TXN-003", "transaction_timeout_limit", "transaction", "事务超时超过 broker 上限，事务生产者将直接失败")
		result.Evidence = evidence
		result.NextActions = []string{"降低 transaction.timeout.ms", "或同步提升 broker transaction.max.timeout.ms", "避免事务生产者配置与 broker 限制不一致"}
		return result
	}

	result := rule.NewPass("TXN-003", "transaction_timeout_limit", "transaction", "事务超时未超过 broker 允许上限")
	result.Evidence = evidence
	return result
}
