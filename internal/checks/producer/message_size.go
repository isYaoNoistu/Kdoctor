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
			result := rule.NewFail("PRD-004", "message_size_budget", "producer", "Kafka logs already show message-too-large failures in the current window")
			result.Evidence = append(evidence, fmt.Sprintf("log_match=%s count=%d", match.ID, match.Count))
			result.NextActions = []string{"compare producer max.request.size with broker message.max.bytes", "reduce payload size or split records before retrying", "check whether compression or schema growth changed message size recently"}
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
		result := rule.NewFail("PRD-004", "message_size_budget", "producer", "configured probe message size is larger than the broker message.max.bytes budget")
		result.Evidence = evidence
		result.NextActions = []string{"reduce probe.message_bytes or producer payload size", "raise broker message.max.bytes only after evaluating replication and memory impact", "keep producer and broker size limits aligned across environments"}
		return result
	}

	if len(evidence) == 0 {
		return rule.NewSkip("PRD-004", "message_size_budget", "producer", "message size budget is not available from the current input set")
	}

	result := rule.NewPass("PRD-004", "message_size_budget", "producer", "no message size budget conflict is visible in the current probe and broker settings")
	result.Evidence = evidence
	return result
}
