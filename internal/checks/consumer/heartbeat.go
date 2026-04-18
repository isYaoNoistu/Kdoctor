package consumer

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type HeartbeatChecker struct {
	SessionTimeoutMs    int
	HeartbeatIntervalMs int
}

func (HeartbeatChecker) ID() string     { return "CSM-004" }
func (HeartbeatChecker) Name() string   { return "heartbeat_session_sanity" }
func (HeartbeatChecker) Module() string { return "consumer" }

func (c HeartbeatChecker) Run(_ context.Context, _ *snapshot.Bundle) model.CheckResult {
	if c.SessionTimeoutMs == 0 && c.HeartbeatIntervalMs == 0 {
		return rule.NewSkip("CSM-004", "heartbeat_session_sanity", "consumer", "当前 profile 未提供 heartbeat/session 参数，暂不评估消费组保活配置")
	}

	evidence := []string{
		fmt.Sprintf("session_timeout_ms=%d", c.SessionTimeoutMs),
		fmt.Sprintf("heartbeat_interval_ms=%d", c.HeartbeatIntervalMs),
	}
	if c.SessionTimeoutMs > 0 && c.HeartbeatIntervalMs > 0 && c.HeartbeatIntervalMs > c.SessionTimeoutMs/3 {
		result := rule.NewWarn("CSM-004", "heartbeat_session_sanity", "consumer", "heartbeat.interval.ms 相对 session.timeout.ms 偏大，消费组保活裕量不足")
		result.Evidence = evidence
		result.NextActions = []string{"让 heartbeat.interval.ms 保持明显小于 session.timeout.ms", "优先采用 heartbeat <= session/3 的保守基线", "如经常 rebalance，请优先核对这里的设置"}
		return result
	}

	result := rule.NewPass("CSM-004", "heartbeat_session_sanity", "consumer", "heartbeat 与 session timeout 组合基本合理")
	result.Evidence = evidence
	return result
}
