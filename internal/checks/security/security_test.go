package security

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestListenerCheckerFailsOnSecurityModeMismatch(t *testing.T) {
	bundle := &snapshot.Bundle{
		Compose: &snapshot.ComposeSnapshot{
			Services: map[string]snapshot.ComposeService{
				"kafka1": {
					Name:  "kafka1",
					Image: "bitnami/kafka:4.0.0",
					Environment: map[string]string{
						"KAFKA_CFG_LISTENERS":                      "INTERNAL://192.168.1.10:9192,EXTERNAL://0.0.0.0:9292",
						"KAFKA_CFG_ADVERTISED_LISTENERS":           "INTERNAL://192.168.1.10:9192,EXTERNAL://203.0.113.10:9292",
						"KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP": "INTERNAL:PLAINTEXT,EXTERNAL:PLAINTEXT",
						"KAFKA_CFG_NODE_ID":                        "1",
						"KAFKA_CFG_PROCESS_ROLES":                  "controller,broker",
					},
				},
			},
		},
	}

	result := ListenerChecker{ExecutionView: "external", SecurityMode: "ssl"}.Run(context.Background(), bundle)
	if result.Status != model.StatusFail {
		t.Fatalf("expected fail, got %s", result.Status)
	}
}

func TestSASLCheckerPassesWhenRequiredMechanismIsEnabled(t *testing.T) {
	bundle := &snapshot.Bundle{
		Compose: &snapshot.ComposeSnapshot{
			Services: map[string]snapshot.ComposeService{
				"kafka1": {
					Name:  "kafka1",
					Image: "bitnami/kafka:4.0.0",
					Environment: map[string]string{
						"KAFKA_CFG_LISTENERS":                      "EXTERNAL://0.0.0.0:9292",
						"KAFKA_CFG_ADVERTISED_LISTENERS":           "EXTERNAL://203.0.113.10:9292",
						"KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP": "EXTERNAL:SASL_SSL",
						"KAFKA_CFG_SASL_ENABLED_MECHANISMS":        "SCRAM-SHA-512,PLAIN",
						"KAFKA_CFG_NODE_ID":                        "1",
						"KAFKA_CFG_PROCESS_ROLES":                  "controller,broker",
					},
				},
			},
		},
	}

	result := SASLChecker{
		ExecutionView: "external",
		SecurityMode:  "sasl_ssl",
		SASLMechanism: "SCRAM-SHA-512",
	}.Run(context.Background(), bundle)
	if result.Status != model.StatusPass {
		t.Fatalf("expected pass, got %s", result.Status)
	}
}
