package docker

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestMountCheckerFailsWhenKafkaDataIsNotMounted(t *testing.T) {
	result := MountChecker{}.Run(context.Background(), &snapshot.Bundle{
		Compose: &snapshot.ComposeSnapshot{
			Services: map[string]snapshot.ComposeService{
				"kafka1": {
					Image:         "bitnami/kafka:4.0.0",
					ContainerName: "kafka1",
					Environment: map[string]string{
						"KAFKA_CFG_NODE_ID":          "1",
						"KAFKA_CFG_PROCESS_ROLES":    "controller,broker",
						"KAFKA_CFG_LOG_DIRS":         "/bitnami/kafka/data",
						"KAFKA_CFG_METADATA_LOG_DIR": "/bitnami/kafka/meta",
					},
				},
			},
		},
		Docker: &snapshot.DockerSnapshot{
			Collected:     true,
			Available:     true,
			ExpectedNames: []string{"kafka1"},
			Containers: []snapshot.DockerContainerStatus{
				{
					Name:    "kafka1",
					Running: true,
					Mounts: []snapshot.DockerMount{
						{Source: "/data/other", Destination: "/tmp"},
					},
				},
			},
		},
	})

	if result.Status != model.StatusFail {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}
}
