package network

import (
	"context"
	"net"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/config"
	"kdoctor/internal/snapshot"
	"kdoctor/internal/transport/tcp"
)

type Collector struct{}

func (Collector) CollectBase(ctx context.Context, env *config.Runtime) *snapshot.NetworkSnapshot {
	out := &snapshot.NetworkSnapshot{}
	for _, address := range bootstrapTargets(env) {
		appendEndpointCheck(ctx, env, &out.BootstrapChecks, "bootstrap", address)
	}
	for _, address := range env.ControllerEndpoints {
		appendEndpointCheck(ctx, env, &out.ControllerChecks, "controller", address)
	}
	return out
}

func bootstrapTargets(env *config.Runtime) []string {
	if env == nil {
		return nil
	}
	if len(env.BootstrapExternal) > 0 {
		return append([]string(nil), env.BootstrapExternal...)
	}
	return append([]string(nil), env.BootstrapInternal...)
}

func (Collector) CollectMetadata(ctx context.Context, env *config.Runtime, network *snapshot.NetworkSnapshot, brokers []snapshot.BrokerSnapshot) *snapshot.NetworkSnapshot {
	if network == nil {
		network = &snapshot.NetworkSnapshot{}
	}
	for _, broker := range brokers {
		appendEndpointCheck(ctx, env, &network.MetadataChecks, "metadata", broker.Address)
	}
	return network
}

func (Collector) CollectComposeControllers(ctx context.Context, env *config.Runtime, network *snapshot.NetworkSnapshot, compose *snapshot.ComposeSnapshot) *snapshot.NetworkSnapshot {
	if network == nil {
		network = &snapshot.NetworkSnapshot{}
	}
	if len(network.ControllerChecks) > 0 || compose == nil {
		return network
	}

	for _, service := range composeutil.KafkaServices(compose) {
		voters, err := composeutil.ParseVoters(service.Environment["KAFKA_CFG_CONTROLLER_QUORUM_VOTERS"])
		if err != nil {
			continue
		}
		for _, address := range voters {
			appendEndpointCheck(ctx, env, &network.ControllerChecks, "controller", address)
		}
	}
	return network
}

func appendEndpointCheck(ctx context.Context, env *config.Runtime, checks *[]snapshot.EndpointCheck, kind string, address string) {
	if checks == nil {
		return
	}
	normalized := normalizeEndpoint(address)
	if normalized == "" || containsEndpoint(*checks, normalized) {
		return
	}

	result := tcp.Dial(ctx, address, env.TCPTimeout)
	*checks = append(*checks, snapshot.EndpointCheck{
		Kind:       kind,
		Address:    normalized,
		Reachable:  result.Reachable,
		DurationMs: result.Duration.Milliseconds(),
		Error:      result.Error,
	})
}

func containsEndpoint(checks []snapshot.EndpointCheck, address string) bool {
	for _, check := range checks {
		if normalizeEndpoint(check.Address) == address {
			return true
		}
	}
	return false
}

func normalizeEndpoint(address string) string {
	address = strings.TrimSpace(address)
	if address == "" {
		return ""
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return strings.ToLower(address)
	}
	return strings.ToLower(net.JoinHostPort(host, port))
}
