package composeutil

import (
	"path/filepath"
	"testing"
)

func TestMapContainerPathToHost(t *testing.T) {
	service := KafkaService{
		Volumes: []string{"./kafka/broker1:/bitnami/kafka"},
	}

	got, ok := MapContainerPathToHost(filepath.Join("workspace", "docker-compose.yml"), service, "/bitnami/kafka/data")
	if !ok {
		t.Fatalf("expected volume mapping to succeed")
	}

	want := filepath.Clean(filepath.Join("workspace", "kafka", "broker1", "data"))
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
