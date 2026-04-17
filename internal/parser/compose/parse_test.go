package compose

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFileSupportsEnvironmentList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "docker-compose.yml")
	content := []byte(`
services:
  kafka1:
    container_name: kafka1
    image: bitnami/kafka:4.0.0
    network_mode: "host"
    mem_limit: 16g
    environment:
      - KAFKA_CFG_NODE_ID=1
      - KAFKA_CFG_PROCESS_ROLES=controller,broker
    volumes:
      - ./kafka/broker1:/bitnami/kafka
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write temp compose: %v", err)
	}

	file, err := ParseFile(path)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}
	if file.Services["kafka1"].Environment["KAFKA_CFG_NODE_ID"] != "1" {
		t.Fatalf("expected environment entry to be parsed, got %#v", file.Services["kafka1"].Environment)
	}
}
