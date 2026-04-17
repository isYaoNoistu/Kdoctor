package lint

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ListenersChecker struct{}

func (ListenersChecker) ID() string     { return "CFG-006" }
func (ListenersChecker) Name() string   { return "listeners_and_advertised_listeners" }
func (ListenersChecker) Module() string { return "lint" }

func (ListenersChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	services := kafkaServices(getCompose(snap))
	if len(services) == 0 {
		return rule.NewSkip("CFG-006", "listeners_and_advertised_listeners", "lint", "compose Kafka services not available")
	}

	usedBindPorts := map[int]string{}
	evidence := []string{}
	for _, service := range services {
		listeners, err := parseListeners(service.Listeners)
		if err != nil {
			result := rule.NewFail("CFG-006", "listeners_and_advertised_listeners", "lint", "listeners format is invalid")
			result.Evidence = []string{fmt.Sprintf("service=%s error=%v", service.ServiceName, err)}
			return result
		}
		advertised, err := parseListeners(service.AdvertisedListeners)
		if err != nil {
			result := rule.NewFail("CFG-006", "listeners_and_advertised_listeners", "lint", "advertised.listeners format is invalid")
			result.Evidence = []string{fmt.Sprintf("service=%s error=%v", service.ServiceName, err)}
			return result
		}
		for _, name := range sortedKeys(listeners) {
			listener := listeners[name]
			if previous, ok := usedBindPorts[listener.Port]; ok {
				result := rule.NewFail("CFG-006", "listeners_and_advertised_listeners", "lint", "listener port conflict detected across Kafka services")
				result.Evidence = []string{
					fmt.Sprintf("port=%d used by %s and %s", listener.Port, previous, service.ServiceName),
				}
				return result
			}
			usedBindPorts[listener.Port] = service.ServiceName
			if name != "CONTROLLER" {
				adv, ok := advertised[name]
				if !ok {
					result := rule.NewFail("CFG-006", "listeners_and_advertised_listeners", "lint", "advertised.listeners is missing a client-facing listener")
					result.Evidence = []string{
						fmt.Sprintf("service=%s missing advertised listener=%s", service.ServiceName, name),
					}
					return result
				}
				if adv.Host == "0.0.0.0" {
					result := rule.NewFail("CFG-006", "listeners_and_advertised_listeners", "lint", "advertised.listeners must not use 0.0.0.0")
					result.Evidence = []string{
						fmt.Sprintf("service=%s listener=%s", service.ServiceName, name),
					}
					return result
				}
			}
			evidence = append(evidence, fmt.Sprintf("%s %s=%s", service.ServiceName, name, listener.Raw))
		}
	}

	result := rule.NewPass("CFG-006", "listeners_and_advertised_listeners", "lint", "listeners and advertised.listeners are structurally consistent")
	result.Evidence = evidence
	return result
}
