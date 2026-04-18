package transaction

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type TopicAbsenceChecker struct {
	TXProbeEnabled  bool
	TransactionalID string
}

func (TopicAbsenceChecker) ID() string     { return "TXN-001" }
func (TopicAbsenceChecker) Name() string   { return "transaction_topic_absence_context" }
func (TopicAbsenceChecker) Module() string { return "transaction" }

func (c TopicAbsenceChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Topic == nil {
		return rule.NewSkip("TXN-001", "transaction_topic_absence_context", "transaction", "当前没有可用的 topic 快照，无法评估事务主题上下文")
	}

	topicPresent := topicExists(bundle, "__transaction_state")
	evidence := []string{
		fmt.Sprintf("tx_probe_enabled=%t", c.TXProbeEnabled),
		fmt.Sprintf("transactional_id_configured=%t", c.TransactionalID != ""),
		fmt.Sprintf("transaction_topic_present=%t", topicPresent),
	}
	if topicPresent {
		result := rule.NewPass("TXN-001", "transaction_topic_absence_context", "transaction", "事务主题已存在")
		result.Evidence = evidence
		return result
	}
	if !c.TXProbeEnabled && c.TransactionalID == "" {
		result := rule.NewWarn("TXN-001", "transaction_topic_absence_context", "transaction", "当前没有看到 __transaction_state，但也没有事务使用证据，这通常只是一条上下文提示")
		result.Evidence = evidence
		result.NextActions = []string{"如果环境未使用事务，这条结果通常可接受", "如准备启用事务，请在上线前确认事务主题可被创建", "避免把未使用事务场景误判成真实故障"}
		return result
	}

	result := rule.NewPass("TXN-001", "transaction_topic_absence_context", "transaction", "事务主题缺失已由更高优先级的事务检查继续评估")
	result.Evidence = evidence
	return result
}
