package kafka

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type InternalTopicsChecker struct{}

func (InternalTopicsChecker) ID() string     { return "KFK-004" }
func (InternalTopicsChecker) Name() string   { return "internal_topics_health" }
func (InternalTopicsChecker) Module() string { return "kafka" }

func (InternalTopicsChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Topic == nil {
		return rule.NewError("KFK-004", "internal_topics_health", "kafka", "internal topics cannot be evaluated", "topic snapshot missing")
	}

	offsets, hasOffsets := findTopic(snap.Topic, "__consumer_offsets")
	txn, hasTxn := findTopic(snap.Topic, "__transaction_state")

	if !hasOffsets {
		result := rule.NewFail("KFK-004", "internal_topics_health", "kafka", "__consumer_offsets topic is missing")
		result.NextActions = []string{"verify cluster metadata integrity", "verify brokers can create and load internal topics", "check controller and broker logs"}
		return result
	}

	offsetsIssues := topicHealthIssues(offsets)
	txnIssues := []string{}
	if hasTxn {
		txnIssues = topicHealthIssues(txn)
	}

	evidence := []string{
		fmt.Sprintf("__consumer_offsets partitions=%d", len(offsets.Partitions)),
	}
	if hasTxn {
		evidence = append(evidence, fmt.Sprintf("__transaction_state partitions=%d", len(txn.Partitions)))
	} else {
		evidence = append(evidence, "__transaction_state missing")
	}
	evidence = append(evidence, offsetsIssues...)
	evidence = append(evidence, txnIssues...)

	if len(offsetsIssues) > 0 || len(txnIssues) > 0 {
		result := rule.NewFail("KFK-004", "internal_topics_health", "kafka", "internal Kafka topics are unhealthy")
		result.Evidence = evidence
		result.NextActions = []string{"verify controller health", "verify broker replication health", "check internal topic leaders and ISR"}
		return result
	}
	if !hasTxn {
		result := rule.NewWarn("KFK-004", "internal_topics_health", "kafka", "__transaction_state topic is not present yet")
		result.Evidence = evidence
		result.NextActions = []string{"this may be acceptable if transactions are unused", "verify transactional producers if they are expected"}
		return result
	}

	result := rule.NewPass("KFK-004", "internal_topics_health", "kafka", "internal Kafka topics are healthy")
	result.Evidence = evidence
	return result
}

func findTopic(topics *snapshot.TopicSnapshot, name string) (snapshot.TopicInfo, bool) {
	if topics == nil {
		return snapshot.TopicInfo{}, false
	}
	for _, topic := range topics.Topics {
		if topic.Name == name {
			return topic, true
		}
	}
	return snapshot.TopicInfo{}, false
}

func topicHealthIssues(topic snapshot.TopicInfo) []string {
	issues := []string{}
	for _, partition := range topic.Partitions {
		if partition.LeaderID == nil {
			issues = append(issues, fmt.Sprintf("%s partition %d has no leader", topic.Name, partition.ID))
			continue
		}
		if len(partition.ISR) < len(partition.Replicas) {
			issues = append(issues, fmt.Sprintf("%s partition %d ISR=%d replicas=%d", topic.Name, partition.ID, len(partition.ISR), len(partition.Replicas)))
		}
	}
	return issues
}
