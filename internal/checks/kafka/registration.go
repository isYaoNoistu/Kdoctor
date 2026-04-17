package kafka

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type RegistrationChecker struct{}

func (RegistrationChecker) ID() string     { return "KFK-002" }
func (RegistrationChecker) Name() string   { return "broker_registration" }
func (RegistrationChecker) Module() string { return "kafka" }

func (RegistrationChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Kafka == nil {
		return rule.NewError("KFK-002", "broker_registration", "kafka", "broker registration cannot be evaluated", "kafka snapshot missing")
	}

	expected := snap.Kafka.ExpectedBrokerCount
	actual := len(snap.Kafka.Brokers)
	result := rule.NewPass("KFK-002", "broker_registration", "kafka", "all expected brokers are registered")
	if expected > 0 {
		result.Evidence = []string{fmt.Sprintf("expected=%d actual=%d", expected, actual)}
	} else {
		result.Evidence = []string{fmt.Sprintf("actual=%d", actual)}
	}
	if expected > 0 && actual < expected {
		result = rule.NewFail("KFK-002", "broker_registration", "kafka", "broker registration count is below expectation")
		result.Evidence = []string{fmt.Sprintf("expected=%d actual=%d", expected, actual)}
		result.NextActions = []string{"verify broker processes are running", "verify node.id and controller quorum settings", "verify broker can register to cluster"}
	}
	return result
}
