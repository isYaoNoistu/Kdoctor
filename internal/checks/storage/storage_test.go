package storage

import (
	"context"
	"testing"

	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func TestLayoutCheckerWarnsWhenMetadataDirIsMissing(t *testing.T) {
	bundle := &snapshot.Bundle{
		Compose: &snapshot.ComposeSnapshot{
			Services: map[string]snapshot.ComposeService{
				"kafka1": {
					Name:  "kafka1",
					Image: "bitnami/kafka:4.0.0",
					Environment: map[string]string{
						"KAFKA_CFG_NODE_ID":       "1",
						"KAFKA_CFG_PROCESS_ROLES": "controller,broker",
						"KAFKA_CFG_LOG_DIRS":      "/bitnami/kafka/data",
					},
				},
			},
		},
	}

	result := LayoutChecker{}.Run(context.Background(), bundle)
	if result.Status != model.StatusWarn {
		t.Fatalf("expected warn, got %s", result.Status)
	}
}

func TestMountPlanningCheckerFailsWithoutVolumeForDataDir(t *testing.T) {
	bundle := &snapshot.Bundle{
		Compose: &snapshot.ComposeSnapshot{
			SourcePath: "/srv/kafka/docker-compose.yml",
			Services: map[string]snapshot.ComposeService{
				"kafka1": {
					Name:    "kafka1",
					Image:   "bitnami/kafka:4.0.0",
					Volumes: []string{"/etc/localtime:/etc/localtime:ro"},
					Environment: map[string]string{
						"KAFKA_CFG_NODE_ID":          "1",
						"KAFKA_CFG_PROCESS_ROLES":    "controller,broker",
						"KAFKA_CFG_LOG_DIRS":         "/bitnami/kafka/data",
						"KAFKA_CFG_METADATA_LOG_DIR": "/bitnami/kafka/meta",
					},
				},
			},
		},
	}

	result := MountPlanningChecker{}.Run(context.Background(), bundle)
	if result.Status != model.StatusFail {
		t.Fatalf("expected fail, got %s", result.Status)
	}
}
