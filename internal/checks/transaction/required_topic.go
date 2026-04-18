package transaction

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type RequiredTopicChecker struct {
	TXProbeEnabled  bool
	TransactionalID string
}

func (RequiredTopicChecker) ID() string     { return "TXN-002" }
func (RequiredTopicChecker) Name() string   { return "transaction_topic_required" }
func (RequiredTopicChecker) Module() string { return "transaction" }

func (c RequiredTopicChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Topic == nil {
		return rule.NewSkip("TXN-002", "transaction_topic_required", "transaction", "当前没有可用的 topic 快照，无法评估事务主题是否必需")
	}

	required := c.TXProbeEnabled || c.TransactionalID != ""
	evidence := []string{
		fmt.Sprintf("tx_probe_enabled=%t", c.TXProbeEnabled),
		fmt.Sprintf("transactional_id=%t", c.TransactionalID != ""),
	}
	if !required {
		return rule.NewSkip("TXN-002", "transaction_topic_required", "transaction", "当前 profile 没有事务使用迹象，暂不把事务主题缺失判成故障")
	}

	if !topicExists(bundle, "__transaction_state") {
		result := rule.NewFail("TXN-002", "transaction_topic_required", "transaction", "环境存在事务使用迹象，但 __transaction_state 缺失")
		result.Evidence = append(evidence, "transaction_topic_present=false")
		result.NextActions = []string{"优先检查 controller 与内部主题创建能力", "确认事务生产者初始化是否已经触发主题创建", "检查 broker 日志中的 transaction coordinator / topic creation 错误"}
		return result
	}

	result := rule.NewPass("TXN-002", "transaction_topic_required", "transaction", "事务使用迹象与事务主题状态一致")
	result.Evidence = append(evidence, "transaction_topic_present=true")
	return result
}
