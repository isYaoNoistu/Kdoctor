package kraft

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestControllerCheckerPassesForPrivateControllerFromExternalView(t *testing.T) {
	controllerID := int32(1)
	result := ControllerChecker{}.Run(context.Background(), &snapshot.Bundle{
		Kafka: &snapshot.KafkaSnapshot{
			ControllerID:      &controllerID,
			ControllerAddress: "192.168.1.10:9193",
			Brokers: []snapshot.BrokerSnapshot{
				{ID: 1, Address: "203.0.113.10:9292"},
			},
		},
		Network: &snapshot.NetworkSnapshot{
			BootstrapChecks: []snapshot.EndpointCheck{
				{Kind: "bootstrap", Address: "203.0.113.10:9292", Reachable: true},
			},
			ControllerChecks: []snapshot.EndpointCheck{
				{Kind: "controller", Address: "192.168.1.10:9193", Reachable: false, Error: "dial timeout"},
			},
		},
	})

	if result.Status != model.StatusPass {
		t.Fatalf("expected PASS, got %s", result.Status)
	}
}
