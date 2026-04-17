package kafka

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestEndpointCheckerFailsOnPrivateBrokerForExternalView(t *testing.T) {
	checker := EndpointChecker{}
	result := checker.Run(context.Background(), &snapshot.Bundle{
		Network: &snapshot.NetworkSnapshot{
			BootstrapChecks: []snapshot.EndpointCheck{
				{Address: "203.0.113.10:9292", Reachable: true},
			},
		},
		Kafka: &snapshot.KafkaSnapshot{
			Brokers: []snapshot.BrokerSnapshot{
				{ID: 1, Address: "192.168.1.10:9092"},
			},
		},
	})

	if result.Status != model.StatusFail {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}
}
