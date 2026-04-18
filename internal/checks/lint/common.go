package lint

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"kdoctor/internal/snapshot"
)

type brokerConfig struct {
	ServiceName              string
	ContainerName            string
	NodeID                   int
	NodeIDRaw                string
	ClusterID                string
	ProcessRoles             []string
	ControllerQuorumVoters   string
	ControllerListenerNames  string
	Listeners                string
	AdvertisedListeners      string
	InterBrokerListenerName  string
	MetadataLogDir           string
	NumPartitions            int
	DefaultReplicationFactor int
	OffsetsReplicationFactor int
	TxnReplicationFactor     int
	TxnMinISR                int
	MinISR                   int
}

type listenerEndpoint struct {
	Name string
	Host string
	Port int
	Raw  string
}

func kafkaServices(compose *snapshot.ComposeSnapshot) []brokerConfig {
	if compose == nil {
		return nil
	}
	names := make([]string, 0, len(compose.Services))
	for name := range compose.Services {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]brokerConfig, 0, len(names))
	for _, name := range names {
		service := compose.Services[name]
		if !isKafkaService(service) {
			continue
		}
		out = append(out, brokerConfig{
			ServiceName:              name,
			ContainerName:            service.ContainerName,
			NodeID:                   mustAtoi(service.Environment["KAFKA_CFG_NODE_ID"]),
			NodeIDRaw:                strings.TrimSpace(service.Environment["KAFKA_CFG_NODE_ID"]),
			ClusterID:                strings.TrimSpace(service.Environment["KAFKA_CLUSTER_ID"]),
			ProcessRoles:             splitCSV(service.Environment["KAFKA_CFG_PROCESS_ROLES"]),
			ControllerQuorumVoters:   strings.TrimSpace(service.Environment["KAFKA_CFG_CONTROLLER_QUORUM_VOTERS"]),
			ControllerListenerNames:  strings.TrimSpace(service.Environment["KAFKA_CFG_CONTROLLER_LISTENER_NAMES"]),
			Listeners:                strings.TrimSpace(service.Environment["KAFKA_CFG_LISTENERS"]),
			AdvertisedListeners:      strings.TrimSpace(service.Environment["KAFKA_CFG_ADVERTISED_LISTENERS"]),
			InterBrokerListenerName:  strings.TrimSpace(service.Environment["KAFKA_CFG_INTER_BROKER_LISTENER_NAME"]),
			MetadataLogDir:           strings.TrimSpace(service.Environment["KAFKA_CFG_METADATA_LOG_DIR"]),
			NumPartitions:            mustAtoi(service.Environment["KAFKA_CFG_NUM_PARTITIONS"]),
			DefaultReplicationFactor: mustAtoi(service.Environment["KAFKA_CFG_DEFAULT_REPLICATION_FACTOR"]),
			OffsetsReplicationFactor: mustAtoi(service.Environment["KAFKA_CFG_OFFSETS_TOPIC_REPLICATION_FACTOR"]),
			TxnReplicationFactor:     mustAtoi(service.Environment["KAFKA_CFG_TRANSACTION_STATE_LOG_REPLICATION_FACTOR"]),
			TxnMinISR:                mustAtoi(service.Environment["KAFKA_CFG_TRANSACTION_STATE_LOG_MIN_ISR"]),
			MinISR:                   mustAtoi(service.Environment["KAFKA_CFG_MIN_INSYNC_REPLICAS"]),
		})
	}
	return out
}

func isKafkaService(service snapshot.ComposeService) bool {
	if strings.Contains(strings.ToLower(service.Image), "kafka") {
		return true
	}
	if _, ok := service.Environment["KAFKA_CFG_NODE_ID"]; ok {
		return true
	}
	if _, ok := service.Environment["KAFKA_CFG_PROCESS_ROLES"]; ok {
		return true
	}
	return false
}

func splitCSV(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func mustAtoi(input string) int {
	v, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return 0
	}
	return v
}

func parseListeners(input string) (map[string]listenerEndpoint, error) {
	out := map[string]listenerEndpoint{}
	for _, item := range splitCSV(input) {
		parts := strings.SplitN(item, "://", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid listener %q", item)
		}
		name := strings.TrimSpace(parts[0])
		host, portStr, err := net.SplitHostPort(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid listener address %q: %w", item, err)
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid listener port %q: %w", item, err)
		}
		out[name] = listenerEndpoint{
			Name: name,
			Host: host,
			Port: port,
			Raw:  item,
		}
	}
	return out, nil
}

func parseVoters(input string) (map[int]string, error) {
	out := map[int]string{}
	for _, item := range splitCSV(input) {
		parts := strings.SplitN(item, "@", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid quorum voter %q", item)
		}
		id, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid voter id %q: %w", item, err)
		}
		out[id] = strings.TrimSpace(parts[1])
	}
	return out, nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
