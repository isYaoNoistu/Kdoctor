package kraft

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestQuorumCheckerSkipsPrivateControllersFromExternalView(t *testing.T) {
	controllerID := int32(1)
	checker := QuorumChecker{}
	result := checker.Run(context.Background(), &snapshot.Bundle{
		Network: &snapshot.NetworkSnapshot{
			BootstrapChecks: []snapshot.EndpointCheck{
				{Address: "203.0.113.10:9292", Reachable: true},
			},
			ControllerChecks: []snapshot.EndpointCheck{
				{Address: "192.168.1.10:9193", Reachable: false, Error: "i/o timeout"},
				{Address: "192.168.1.10:9195", Reachable: false, Error: "i/o timeout"},
				{Address: "192.168.1.10:9197", Reachable: false, Error: "i/o timeout"},
			},
		},
		Kafka: &snapshot.KafkaSnapshot{
			ControllerID: &controllerID,
		},
	})

	if result.Status != model.StatusSkip {
		t.Fatalf("expected status SKIP, got %s", result.Status)
	}
}

func TestQuorumCheckerReturnsCritWhenMajorityActuallyLost(t *testing.T) {
	checker := QuorumChecker{}
	result := checker.Run(context.Background(), &snapshot.Bundle{
		Network: &snapshot.NetworkSnapshot{
			BootstrapChecks: []snapshot.EndpointCheck{
				{Address: "10.0.0.11:9192", Reachable: true},
			},
			ControllerChecks: []snapshot.EndpointCheck{
				{Address: "10.0.0.11:9193", Reachable: false, Error: "i/o timeout"},
				{Address: "10.0.0.12:9193", Reachable: false, Error: "i/o timeout"},
				{Address: "10.0.0.13:9193", Reachable: false, Error: "i/o timeout"},
			},
		},
	})

	if result.Status != model.StatusCrit {
		t.Fatalf("expected status CRIT, got %s", result.Status)
	}
}

func TestQuorumCheckerSkipsWhenControllerEndpointsAreUnavailable(t *testing.T) {
	checker := QuorumChecker{}
	result := checker.Run(context.Background(), &snapshot.Bundle{})
	if result.Status != model.StatusSkip {
		t.Fatalf("expected status SKIP, got %s", result.Status)
	}
}
