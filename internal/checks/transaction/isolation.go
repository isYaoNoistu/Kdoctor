package transaction

import (
	"context"
	"fmt"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type IsolationChecker struct {
	IsolationLevel string
	TXProbeEnabled bool
}

func (IsolationChecker) ID() string     { return "TXN-004" }
func (IsolationChecker) Name() string   { return "read_committed_visibility" }
func (IsolationChecker) Module() string { return "transaction" }

func (c IsolationChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	level := strings.ToLower(strings.TrimSpace(c.IsolationLevel))
	if level == "" {
		return rule.NewSkip("TXN-004", "read_committed_visibility", "transaction", "当前未配置 isolation.level，暂不评估 read_committed 语义")
	}

	evidence := []string{
		fmt.Sprintf("isolation_level=%s", level),
		fmt.Sprintf("tx_probe_enabled=%t", c.TXProbeEnabled),
	}
	if level != "read_committed" {
		result := rule.NewPass("TXN-004", "read_committed_visibility", "transaction", "当前消费者不走 read_committed 语义")
		result.Evidence = evidence
		return result
	}
	if !c.TXProbeEnabled {
		result := rule.NewWarn("TXN-004", "read_committed_visibility", "transaction", "消费者使用 read_committed，但当前未启用事务探针，无法进一步验证事务提交可见性")
		result.Evidence = evidence
		result.NextActions = []string{"在 full/incident 模式开启事务探针", "结合 __transaction_state 与事务日志继续确认", "避免把 read_committed 下的可见性差异误判成普通消费问题"}
		return result
	}
	if bundle != nil && !topicExists(bundle, "__transaction_state") {
		result := rule.NewFail("TXN-004", "read_committed_visibility", "transaction", "消费者使用 read_committed，但事务主题缺失，事务可见性不可信")
		result.Evidence = append(evidence, "transaction_topic_present=false")
		result.NextActions = []string{"优先修复 __transaction_state", "确认事务 coordinator 正常工作", "再继续验证 read_committed 的可见性"}
		return result
	}

	result := rule.NewPass("TXN-004", "read_committed_visibility", "transaction", "read_committed 语义已具备基本前提")
	result.Evidence = evidence
	return result
}
