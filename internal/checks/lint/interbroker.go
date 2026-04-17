package lint

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type InterBrokerListenerChecker struct{}

func (InterBrokerListenerChecker) ID() string     { return "CFG-007" }
func (InterBrokerListenerChecker) Name() string   { return "inter_broker_listener_name" }
func (InterBrokerListenerChecker) Module() string { return "lint" }

func (InterBrokerListenerChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	services := kafkaServices(getCompose(snap))
	if len(services) == 0 {
		return rule.NewSkip("CFG-007", "inter_broker_listener_name", "lint", "compose Kafka services not available")
	}

	evidence := []string{}
	for _, service := range services {
		if service.InterBrokerListenerName == "" {
			result := rule.NewFail("CFG-007", "inter_broker_listener_name", "lint", "inter.broker.listener.name is missing")
			result.Evidence = []string{fmt.Sprintf("service=%s", service.ServiceName)}
			return result
		}
		listeners, err := parseListeners(service.Listeners)
		if err != nil {
			result := rule.NewFail("CFG-007", "inter_broker_listener_name", "lint", "listeners format is invalid while validating inter.broker.listener.name")
			result.Evidence = []string{fmt.Sprintf("service=%s error=%v", service.ServiceName, err)}
			return result
		}
		if _, ok := listeners[service.InterBrokerListenerName]; !ok {
			result := rule.NewFail("CFG-007", "inter_broker_listener_name", "lint", "inter.broker.listener.name does not exist in listeners")
			result.Evidence = []string{
				fmt.Sprintf("service=%s listener=%s", service.ServiceName, service.InterBrokerListenerName),
			}
			return result
		}
		if service.InterBrokerListenerName == "EXTERNAL" {
			result := rule.NewWarn("CFG-007", "inter_broker_listener_name", "lint", "inter.broker.listener.name points to EXTERNAL listener")
			result.Evidence = []string{
				fmt.Sprintf("service=%s listener=%s", service.ServiceName, service.InterBrokerListenerName),
			}
			result.NextActions = []string{"prefer INTERNAL listener for broker-to-broker traffic", "verify listener security and routing for inter-broker communication"}
			return result
		}
		evidence = append(evidence, fmt.Sprintf("%s inter.broker.listener.name=%s", service.ServiceName, service.InterBrokerListenerName))
	}

	result := rule.NewPass("CFG-007", "inter_broker_listener_name", "lint", "inter.broker.listener.name points to valid listener")
	result.Evidence = evidence
	return result
}
