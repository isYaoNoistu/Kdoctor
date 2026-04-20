package kraft

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestEndpointConfigCheckerUsesControllerNodeAndListenerInsteadOfMetadataBrokerAddress(t *testing.T) {
	controllerID := int32(2)
	result := EndpointConfigChecker{}.Run(context.Background(), &snapshot.Bundle{
		Kafka: &snapshot.KafkaSnapshot{
			ControllerID:      &controllerID,
			ControllerAddress: "192.168.119.7:9194",
		},
		Compose: &snapshot.ComposeSnapshot{
			Services: map[string]snapshot.ComposeService{
				"kafka1": {
					Name: "kafka1",
					Environment: map[string]string{
						"KAFKA_CFG_NODE_ID":                   "1",
						"KAFKA_CFG_PROCESS_ROLES":             "controller,broker",
						"KAFKA_CFG_LISTENERS":                 "INTERNAL://192.168.119.7:9192,CONTROLLER://192.168.119.7:9193",
						"KAFKA_CFG_CONTROLLER_LISTENER_NAMES": "CONTROLLER",
						"KAFKA_CFG_CONTROLLER_QUORUM_VOTERS":  "1@192.168.119.7:9193,2@192.168.119.7:9195,3@192.168.119.7:9197",
					},
				},
				"kafka2": {
					Name: "kafka2",
					Environment: map[string]string{
						"KAFKA_CFG_NODE_ID":                   "2",
						"KAFKA_CFG_PROCESS_ROLES":             "controller,broker",
						"KAFKA_CFG_LISTENERS":                 "INTERNAL://192.168.119.7:9194,CONTROLLER://192.168.119.7:9195",
						"KAFKA_CFG_CONTROLLER_LISTENER_NAMES": "CONTROLLER",
						"KAFKA_CFG_CONTROLLER_QUORUM_VOTERS":  "1@192.168.119.7:9193,2@192.168.119.7:9195,3@192.168.119.7:9197",
					},
				},
			},
		},
	})

	if result.Status != model.StatusPass {
		t.Fatalf("expected PASS, got %s", result.Status)
	}
}
