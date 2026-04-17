package lint

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ReplicationChecker struct {
	ExpectedBrokerCount int
}

func (ReplicationChecker) ID() string     { return "CFG-008" }
func (ReplicationChecker) Name() string   { return "replication_and_isr_legality" }
func (ReplicationChecker) Module() string { return "lint" }

func (c ReplicationChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	services := kafkaServices(getCompose(snap))
	if len(services) == 0 {
		return rule.NewSkip("CFG-008", "replication_and_isr_legality", "lint", "compose Kafka services not available")
	}
	if c.ExpectedBrokerCount <= 0 {
		c.ExpectedBrokerCount = len(services)
	}

	broker := services[0]
	evidence := []string{
		fmt.Sprintf("default.replication.factor=%d", broker.DefaultReplicationFactor),
		fmt.Sprintf("offsets.topic.replication.factor=%d", broker.OffsetsReplicationFactor),
		fmt.Sprintf("transaction.state.log.replication.factor=%d", broker.TxnReplicationFactor),
		fmt.Sprintf("transaction.state.log.min.isr=%d", broker.TxnMinISR),
		fmt.Sprintf("min.insync.replicas=%d", broker.MinISR),
		fmt.Sprintf("broker_count=%d", c.ExpectedBrokerCount),
	}

	fail := func(summary string) model.CheckResult {
		result := rule.NewFail("CFG-008", "replication_and_isr_legality", "lint", summary)
		result.Evidence = evidence
		return result
	}

	if broker.DefaultReplicationFactor > c.ExpectedBrokerCount {
		return fail("default.replication.factor is greater than broker count")
	}
	if broker.OffsetsReplicationFactor > c.ExpectedBrokerCount {
		return fail("offsets.topic.replication.factor is greater than broker count")
	}
	if broker.TxnReplicationFactor > c.ExpectedBrokerCount {
		return fail("transaction.state.log.replication.factor is greater than broker count")
	}
	if broker.TxnMinISR > broker.TxnReplicationFactor {
		return fail("transaction.state.log.min.isr is greater than transaction.state.log.replication.factor")
	}
	if broker.MinISR > broker.DefaultReplicationFactor {
		return fail("min.insync.replicas is greater than default.replication.factor")
	}
	if broker.MinISR == broker.DefaultReplicationFactor {
		result := rule.NewWarn("CFG-008", "replication_and_isr_legality", "lint", "min.insync.replicas equals default.replication.factor and leaves no slack")
		result.Evidence = evidence
		result.NextActions = []string{"verify this is intentional", "ensure producer ack expectations match the strict ISR policy"}
		return result
	}

	result := rule.NewPass("CFG-008", "replication_and_isr_legality", "lint", "replication and ISR settings are structurally legal")
	result.Evidence = evidence
	return result
}
