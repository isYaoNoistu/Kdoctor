package consumer

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type PollIntervalChecker struct {
	MaxPollIntervalMs int
	SessionTimeoutMs  int
}

func (PollIntervalChecker) ID() string     { return "CSM-003" }
func (PollIntervalChecker) Name() string   { return "max_poll_interval_sanity" }
func (PollIntervalChecker) Module() string { return "consumer" }

func (c PollIntervalChecker) Run(_ context.Context, _ *snapshot.Bundle) model.CheckResult {
	if c.MaxPollIntervalMs == 0 {
		return rule.NewSkip("CSM-003", "max_poll_interval_sanity", "consumer", "当前 profile 未提供 max.poll.interval.ms，暂不评估消费处理窗口")
	}

	evidence := []string{fmt.Sprintf("max_poll_interval_ms=%d", c.MaxPollIntervalMs)}
	if c.SessionTimeoutMs > 0 {
		evidence = append(evidence, fmt.Sprintf("session_timeout_ms=%d", c.SessionTimeoutMs))
	}

	result := rule.NewPass("CSM-003", "max_poll_interval_sanity", "consumer", "max.poll.interval.ms 已显式配置")
	result.Evidence = evidence
	if c.MaxPollIntervalMs > 0 && c.MaxPollIntervalMs < 60000 {
		result = rule.NewWarn("CSM-003", "max_poll_interval_sanity", "consumer", "max.poll.interval.ms 偏短，业务处理稍慢时就可能被踢出消费组")
		result.Evidence = evidence
		result.NextActions = []string{"结合业务处理时长评估 max.poll.interval.ms", "避免处理批次时间经常接近 poll 上限", "如存在 rebalance 风暴，请优先核对这里的设置"}
	}
	return result
}
