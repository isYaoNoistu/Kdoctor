package lint

import (
	"context"
	"fmt"
	"sort"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type BrokerIdentityChecker struct{}

func (BrokerIdentityChecker) ID() string     { return "CFG-011" }
func (BrokerIdentityChecker) Name() string   { return "broker_identity_uniqueness" }
func (BrokerIdentityChecker) Module() string { return "lint" }

func (BrokerIdentityChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	services := kafkaServices(getCompose(snap))
	if len(services) == 0 {
		return rule.NewSkip("CFG-011", "broker_identity_uniqueness", "lint", "compose Kafka services not available")
	}

	nodeToService := map[int]string{}
	addressToServices := map[string][]string{}
	evidence := []string{}
	for _, service := range services {
		if service.NodeID > 0 {
			if previous, ok := nodeToService[service.NodeID]; ok {
				result := rule.NewFail("CFG-011", "broker_identity_uniqueness", "lint", "duplicate node.id detected in compose Kafka services")
				result.Evidence = []string{fmt.Sprintf("node_id=%d services=%s,%s", service.NodeID, previous, service.ServiceName)}
				result.NextActions = []string{"assign unique node.id values to every broker", "check whether stale compose fragments were merged together", "verify broker IDs match the intended deployment topology"}
				return result
			}
			nodeToService[service.NodeID] = service.ServiceName
		}

		advertised, err := parseListeners(service.AdvertisedListeners)
		if err != nil {
			continue
		}
		for _, name := range sortedKeys(advertised) {
			listener := advertised[name]
			address := fmt.Sprintf("%s:%d", listener.Host, listener.Port)
			addressToServices[address] = append(addressToServices[address], service.ServiceName)
			evidence = append(evidence, fmt.Sprintf("service=%s listener=%s address=%s", service.ServiceName, name, address))
		}
	}

	duplicates := []string{}
	for address, services := range addressToServices {
		if len(services) <= 1 {
			continue
		}
		sort.Strings(services)
		duplicates = append(duplicates, fmt.Sprintf("address=%s services=%v", address, services))
	}

	result := rule.NewPass("CFG-011", "broker_identity_uniqueness", "lint", "node IDs and advertised broker addresses are unique in compose")
	result.Evidence = evidence
	if len(duplicates) > 0 {
		result = rule.NewFail("CFG-011", "broker_identity_uniqueness", "lint", "multiple Kafka services advertise the same broker address")
		result.Evidence = append(result.Evidence, duplicates...)
		result.NextActions = []string{"assign unique advertised listener addresses per broker", "avoid reusing the same external port across multiple host-network brokers", "keep broker identity stable across compose, metadata, and client expectations"}
	}
	return result
}
