package lint

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestListenersCheckerAcceptsCurrentStyleConfig(t *testing.T) {
	checker := ListenersChecker{}
	result := checker.Run(context.Background(), &snapshot.Bundle{
		Compose: &snapshot.ComposeSnapshot{
			Services: map[string]snapshot.ComposeService{
				"kafka1": {
					Name:  "kafka1",
					Image: "bitnami/kafka:4.0.0",
					Environment: map[string]string{
						"KAFKA_CFG_NODE_ID":                  "1",
						"KAFKA_CFG_LISTENERS":                "INTERNAL://192.168.1.10:9192,EXTERNAL://0.0.0.0:9292,CONTROLLER://192.168.1.10:9193",
						"KAFKA_CFG_ADVERTISED_LISTENERS":     "INTERNAL://192.168.1.10:9192,EXTERNAL://203.0.113.10:9292",
						"KAFKA_CFG_PROCESS_ROLES":            "controller,broker",
						"KAFKA_CFG_CONTROLLER_QUORUM_VOTERS": "1@192.168.1.10:9193",
					},
				},
			},
		},
	})

	if result.Status != model.StatusPass {
		t.Fatalf("expected PASS, got %s", result.Status)
	}
}

func TestListenersCheckerRejectsAdvertisedZeroAddress(t *testing.T) {
	checker := ListenersChecker{}
	result := checker.Run(context.Background(), &snapshot.Bundle{
		Compose: &snapshot.ComposeSnapshot{
			Services: map[string]snapshot.ComposeService{
				"kafka1": {
					Name:  "kafka1",
					Image: "bitnami/kafka:4.0.0",
					Environment: map[string]string{
						"KAFKA_CFG_NODE_ID":              "1",
						"KAFKA_CFG_LISTENERS":            "INTERNAL://192.168.1.10:9192,EXTERNAL://0.0.0.0:9292",
						"KAFKA_CFG_ADVERTISED_LISTENERS": "INTERNAL://192.168.1.10:9192,EXTERNAL://0.0.0.0:9292",
						"KAFKA_CFG_PROCESS_ROLES":        "controller,broker",
					},
				},
			},
		},
	})

	if result.Status != model.StatusFail {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}
}
