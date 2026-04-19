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

type MessageSizeChecker struct {
	ProbeMessageBytes int
}

func (MessageSizeChecker) ID() string     { return "PRD-004" }
func (MessageSizeChecker) Name() string   { return "message_size_budget" }
func (MessageSizeChecker) Module() string { return "producer" }

func (c MessageSizeChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	evidence := []string{}
	if c.ProbeMessageBytes > 0 {
		evidence = append(evidence, fmt.Sprintf("probe_message_bytes=%d", c.ProbeMessageBytes))
	}

	if bundle != nil && bundle.Logs != nil {
		for _, match := range bundle.Logs.Matches {
			if match.ID != "LOG-MESSAGE-TOO-LARGE" {
				continue
			}
			result := rule.NewFail("PRD-004", "message_size_budget", "producer", "当前时间窗内的 Kafka 日志已经出现 message-too-large 失败")
			result.Evidence = append(evidence, fmt.Sprintf("log_match=%s count=%d", match.ID, match.Count))
			result.NextActions = []string{"对比 producer max.request.size 与 broker message.max.bytes", "在重试前减小消息体或拆分记录", "检查压缩方式或 schema 增长是否导致消息变大"}
			return result
		}
	}

	minBrokerLimit := int64(0)
	if bundle != nil && bundle.Compose != nil {
		for _, service := range composeutil.KafkaServices(bundle.Compose) {
			raw := strings.TrimSpace(service.Environment["KAFKA_CFG_MESSAGE_MAX_BYTES"])
			if raw == "" {
				continue
			}
			value, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || value <= 0 {
				continue
			}
			evidence = append(evidence, fmt.Sprintf("service=%s message_max_bytes=%d", service.ServiceName, value))
			if minBrokerLimit == 0 || value < minBrokerLimit {
				minBrokerLimit = value
			}
		}
	}

	if c.ProbeMessageBytes > 0 && minBrokerLimit > 0 && int64(c.ProbeMessageBytes) > minBrokerLimit {
		result := rule.NewFail("PRD-004", "message_size_budget", "producer", "当前探针消息大小已经超过 broker message.max.bytes 限制")
		result.Evidence = evidence
		result.NextActions = []string{"降低 probe.message_bytes 或生产消息体大小", "只有在评估复制和内存影响后再提高 broker message.max.bytes", "保持 producer 与 broker 的大小限制在各环境中一致"}
		return result
	}

	if len(evidence) == 0 {
		return rule.NewSkip("PRD-004", "message_size_budget", "producer", "当前输入下没有可用的消息大小预算信息")
	}

	result := rule.NewPass("PRD-004", "message_size_budget", "producer", "当前探针与 broker 配置未见消息大小预算冲突")
	result.Evidence = evidence
	return result
}
