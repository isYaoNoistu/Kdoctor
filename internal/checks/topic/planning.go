package topic

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type PlanningChecker struct {
	ExpectedBrokerCount int
}

func (PlanningChecker) ID() string     { return "TOP-011" }
func (PlanningChecker) Name() string   { return "topic_planning" }
func (PlanningChecker) Module() string { return "topic" }

func (c PlanningChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Topic == nil {
		return rule.NewSkip("TOP-011", "topic_planning", "topic", "当前没有可用的 topic 快照，无法评估 topic 规划")
	}
	if c.ExpectedBrokerCount <= 0 && bundle.Kafka != nil {
		c.ExpectedBrokerCount = bundle.Kafka.ExpectedBrokerCount
		if c.ExpectedBrokerCount == 0 {
			c.ExpectedBrokerCount = len(bundle.Kafka.Brokers)
		}
	}
	if c.ExpectedBrokerCount <= 0 {
		return rule.NewSkip("TOP-011", "topic_planning", "topic", "当前没有可用的 broker 数量基线，暂不评估 topic 规划")
	}

	failures := 0
	warnings := 0
	failEvidence := []string{}
	warnEvidence := []string{}
	for _, topic := range bundle.Topic.Topics {
		if len(topic.Partitions) == 0 {
			continue
		}
		replicationFactor := 0
		if len(topic.Partitions[0].Replicas) > 0 {
			replicationFactor = len(topic.Partitions[0].Replicas)
		}
		partitionCount := len(topic.Partitions)

		switch {
		case replicationFactor > c.ExpectedBrokerCount:
			failures++
			failEvidence = append(failEvidence, fmt.Sprintf("主题=%s 分区数=%d 副本因子=%d broker 数=%d", topic.Name, partitionCount, replicationFactor, c.ExpectedBrokerCount))
		case partitionCount < c.ExpectedBrokerCount:
			warnings++
			warnEvidence = append(warnEvidence, fmt.Sprintf("主题=%s 分区数=%d broker 数=%d", topic.Name, partitionCount, c.ExpectedBrokerCount))
		}
	}

	result := rule.NewPass("TOP-011", "topic_planning", "topic", "topic 分区与副本规划在当前 broker 数下结构合理")
	if failures > 0 {
		result = rule.NewFail("TOP-011", "topic_planning", "topic", "存在 topic 副本因子高于 broker 数，属于明显规划错误")
		result.Evidence = limitPlanningEvidence(failEvidence, planningEvidenceLimit)
		result.NextActions = []string{"调整 topic replication factor 或增加 broker 数", "检查默认 RF 与业务 topic 的建表参数", "避免 RF 高于当前可用 broker 数"}
		return result
	}
	if warnings > 0 {
		result = rule.NewWarn("TOP-011", "topic_planning", "topic", "部分 topic 分区数低于 broker 数，分布与扩展性可能不理想")
		result.Evidence = limitPlanningEvidence(warnEvidence, planningEvidenceLimit)
		result.NextActions = []string{"评估业务 topic 的分区数是否过少", "关注 leader 分布与单 broker 热点风险", "在扩容或重平衡前复核 topic 规划"}
	}
	return result
}

const planningEvidenceLimit = 20

func limitPlanningEvidence(items []string, maxItems int) []string {
	if maxItems <= 0 || len(items) <= maxItems {
		return items
	}
	trimmed := append([]string(nil), items[:maxItems]...)
	trimmed = append(trimmed, fmt.Sprintf("其余 %d 个命中 topic 已省略", len(items)-maxItems))
	return trimmed
}
