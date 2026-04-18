package producer

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

type TxTimeoutChecker struct {
	TransactionTimeoutMs int
}

func (TxTimeoutChecker) ID() string     { return "PRD-006" }
func (TxTimeoutChecker) Name() string   { return "transaction_timeout_sanity" }
func (TxTimeoutChecker) Module() string { return "producer" }

func (c TxTimeoutChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if c.TransactionTimeoutMs == 0 {
		return rule.NewSkip("PRD-006", "transaction_timeout_sanity", "producer", "当前 profile 未提供 transaction timeout，暂不评估事务超时上限")
	}

	evidence := []string{fmt.Sprintf("transaction_timeout_ms=%d", c.TransactionTimeoutMs)}
	maxBrokerTimeout := 0
	if bundle != nil && bundle.Compose != nil {
		for _, service := range composeutil.KafkaServices(bundle.Compose) {
			if raw := strings.TrimSpace(service.Environment["KAFKA_CFG_TRANSACTION_MAX_TIMEOUT_MS"]); raw != "" {
				if value, err := strconv.Atoi(raw); err == nil {
					if maxBrokerTimeout == 0 || value < maxBrokerTimeout {
						maxBrokerTimeout = value
					}
					evidence = append(evidence, fmt.Sprintf("service=%s broker_transaction_max_timeout_ms=%d", service.ServiceName, value))
				}
			}
		}
	}
	if maxBrokerTimeout > 0 && c.TransactionTimeoutMs > maxBrokerTimeout {
		result := rule.NewFail("PRD-006", "transaction_timeout_sanity", "producer", "transaction.timeout.ms 高于 broker 允许上限，事务生产者会直接报错")
		result.Evidence = evidence
		result.NextActions = []string{"降低 producer transaction.timeout.ms", "或同步提升 broker transaction.max.timeout.ms", "避免事务配置与 broker 限制脱节"}
		return result
	}

	result := rule.NewPass("PRD-006", "transaction_timeout_sanity", "producer", "事务超时配置未超过 broker 允许上限")
	result.Evidence = evidence
	return result
}
