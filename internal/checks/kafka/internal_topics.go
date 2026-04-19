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
		evidence := []string{}
		if snap.Probe != nil {
			evidence = append(evidence,
				fmt.Sprintf("commit_executed=%t", snap.Probe.CommitExecuted),
				fmt.Sprintf("commit_ok=%t", snap.Probe.CommitOK),
			)
			if snap.Probe.FailureStage != "" {
				evidence = append(evidence, fmt.Sprintf("failure_stage=%s", snap.Probe.FailureStage))
			}
		}

		if snap.Probe != nil && snap.Probe.CommitExecuted {
			result := rule.NewFail("KFK-004", "internal_topics_health", "kafka", "__consumer_offsets is missing after consumer group commit probe executed")
			result.Evidence = evidence
			result.NextActions = []string{"verify cluster metadata integrity", "verify brokers can create and load internal topics", "check controller and broker logs"}
			return result
		}

		result := rule.NewWarn("KFK-004", "internal_topics_health", "kafka", "__consumer_offsets is not present yet; cluster may still be fresh or commit path has not run")
		result.Evidence = evidence
		result.NextActions = []string{"verify cluster metadata integrity", "verify brokers can create and load internal topics", "check controller and broker logs"}
		return result
	}

	offsetsIssues := topicHealthIssues(offsets)
	txnIssues := []string{}
	if hasTxn {
		txnIssues = topicHealthIssues(txn)
	}

	evidence := []string{
		fmt.Sprintf("__consumer_offsets 分区数=%d", len(offsets.Partitions)),
	}
	if hasTxn {
		evidence = append(evidence, fmt.Sprintf("__transaction_state 分区数=%d", len(txn.Partitions)))
	} else {
		evidence = append(evidence, "__transaction_state 未出现")
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
		summary := "__transaction_state topic is not present yet"
		if snap.TransactionExpected {
			summary = "__transaction_state topic is not present yet; transaction-specific checks will continue the assessment"
		} else {
			summary = "__transaction_state topic is not present, but the current run has no transaction usage evidence"
		}
		result := rule.NewPass("KFK-004", "internal_topics_health", "kafka", summary)
		result.Evidence = evidence
		if snap.TransactionExpected {
			result.NextActions = []string{"review transaction-specific checks such as TXN-001/TXN-002", "verify transactional producers if they are expected"}
		} else {
			result.NextActions = []string{"this may be acceptable if transactions are unused", "verify transactional producers if they are expected"}
		}
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
			issues = append(issues, fmt.Sprintf("主题=%s 分区=%d 没有 leader", topic.Name, partition.ID))
			continue
		}
		if len(partition.ISR) < len(partition.Replicas) {
			issues = append(issues, fmt.Sprintf("主题=%s 分区=%d ISR=%d 副本数=%d", topic.Name, partition.ID, len(partition.ISR), len(partition.Replicas)))
		}
	}
	return issues
}
