package kafka

import (
	"context"
	"fmt"
	"net"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type EndpointChecker struct{}

func (EndpointChecker) ID() string     { return "KFK-003" }
func (EndpointChecker) Name() string   { return "broker_endpoint_legality" }
func (EndpointChecker) Module() string { return "kafka" }

func (EndpointChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	if snap == nil || snap.Kafka == nil {
		return rule.NewError("KFK-003", "broker_endpoint_legality", "kafka", "broker endpoint legality cannot be evaluated", "kafka snapshot missing")
	}
	if len(snap.Kafka.Brokers) == 0 {
		return rule.NewFail("KFK-003", "broker_endpoint_legality", "kafka", "no broker endpoints were returned by metadata")
	}

	evidence := make([]string, 0, len(snap.Kafka.Brokers))
	seen := map[string]int32{}
	externalView := isExternalBootstrapView(snap)
	privateInExternalView := []string{}

	for _, broker := range snap.Kafka.Brokers {
		if _, _, err := net.SplitHostPort(broker.Address); err != nil {
			result := rule.NewFail("KFK-003", "broker_endpoint_legality", "kafka", "metadata returned malformed broker endpoint")
			result.Evidence = []string{fmt.Sprintf("broker_id=%d address=%s error=%v", broker.ID, broker.Address, err)}
			return result
		}
		if previous, ok := seen[broker.Address]; ok {
			result := rule.NewFail("KFK-003", "broker_endpoint_legality", "kafka", "metadata returned duplicate broker endpoints")
			result.Evidence = []string{
				fmt.Sprintf("duplicate address=%s broker_ids=%d,%d", broker.Address, previous, broker.ID),
			}
			return result
		}
		seen[broker.Address] = broker.ID
		evidence = append(evidence, fmt.Sprintf("broker_id=%d address=%s", broker.ID, broker.Address))
		if externalView && isPrivateEndpoint(broker.Address) {
			privateInExternalView = append(privateInExternalView, broker.Address)
		}
	}

	result := rule.NewPass("KFK-003", "broker_endpoint_legality", "kafka", "broker endpoints are structurally valid")
	result.Evidence = evidence
	if len(privateInExternalView) > 0 {
		result = rule.NewFail("KFK-003", "broker_endpoint_legality", "kafka", "metadata returned private broker endpoints for the current external client view")
		result.Evidence = append(evidence, fmt.Sprintf("private_endpoints=%v", privateInExternalView))
		result.NextActions = []string{"verify advertised.listeners for external clients", "ensure metadata returns routable addresses for the current client network"}
	}
	return result
}
