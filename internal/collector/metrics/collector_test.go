package metrics

import (
	"strings"
	"testing"
)

func TestParsePrometheus(t *testing.T) {
	input := `
# HELP kafka_server_replicamanager_underreplicatedpartitions
kafka_server_replicamanager_underreplicatedpartitions 2
kafka_network_socketserver_networkprocessoravgidlepercent{listener="INTERNAL"} 0.25
`

	values, err := parsePrometheus(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse prometheus: %v", err)
	}
	if values["kafka_server_replicamanager_underreplicatedpartitions"] != 2 {
		t.Fatalf("unexpected under replicated value: %v", values["kafka_server_replicamanager_underreplicatedpartitions"])
	}
	if values["kafka_network_socketserver_networkprocessoravgidlepercent"] != 0.25 {
		t.Fatalf("unexpected network idle value: %v", values["kafka_network_socketserver_networkprocessoravgidlepercent"])
	}
}
