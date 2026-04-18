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
		return rule.NewSkip("TOP-011", "topic_planning", "topic", "当前没有可用的 broker 数基线，暂不评估 topic 规划")
	}

	failures := 0
	warnings := 0
	evidence := []string{}
	for _, topic := range bundle.Topic.Topics {
		if len(topic.Partitions) == 0 {
			continue
		}
		replicationFactor := 0
		if len(topic.Partitions[0].Replicas) > 0 {
			replicationFactor = len(topic.Partitions[0].Replicas)
		}
		partitionCount := len(topic.Partitions)
		evidence = append(evidence, fmt.Sprintf("topic=%s partitions=%d rf=%d", topic.Name, partitionCount, replicationFactor))
		if replicationFactor > c.ExpectedBrokerCount {
			failures++
		} else if partitionCount < c.ExpectedBrokerCount {
			warnings++
		}
	}

	result := rule.NewPass("TOP-011", "topic_planning", "topic", "topic 分区与副本规划在当前 broker 数下结构合理")
	result.Evidence = evidence
	if failures > 0 {
		result = rule.NewFail("TOP-011", "topic_planning", "topic", "存在 topic 副本因子高于 broker 数，属于明显规划错误")
		result.Evidence = evidence
		result.NextActions = []string{"调整 topic replication factor 或增加 broker 数", "检查默认 RF 与业务 topic 的建表参数", "避免 RF 高于当前可用 broker 数"}
		return result
	}
	if warnings > 0 {
		result = rule.NewWarn("TOP-011", "topic_planning", "topic", "部分 topic 分区数低于 broker 数，分布与扩展性可能不理想")
		result.Evidence = evidence
		result.NextActions = []string{"评估业务 topic 的 partitions 是否过少", "关注 leader 分布与单 broker 热点风险", "在扩容或重平衡前复核 topic 规划"}
	}
	return result
}
