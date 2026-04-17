package network

import (
	"context"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/config"
	"kdoctor/internal/snapshot"
	"kdoctor/internal/transport/tcp"
)

type Collector struct{}

func (Collector) CollectBase(ctx context.Context, env *config.Runtime) *snapshot.NetworkSnapshot {
	out := &snapshot.NetworkSnapshot{}
	for _, address := range bootstrapTargets(env) {
		result := tcp.Dial(ctx, address, env.TCPTimeout)
		out.BootstrapChecks = append(out.BootstrapChecks, snapshot.EndpointCheck{
			Kind:       "bootstrap",
			Address:    address,
			Reachable:  result.Reachable,
			DurationMs: result.Duration.Milliseconds(),
			Error:      result.Error,
		})
	}
	for _, address := range env.ControllerEndpoints {
		result := tcp.Dial(ctx, address, env.TCPTimeout)
		out.ControllerChecks = append(out.ControllerChecks, snapshot.EndpointCheck{
			Kind:       "controller",
			Address:    address,
			Reachable:  result.Reachable,
			DurationMs: result.Duration.Milliseconds(),
			Error:      result.Error,
		})
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
		result := tcp.Dial(ctx, broker.Address, env.TCPTimeout)
		network.MetadataChecks = append(network.MetadataChecks, snapshot.EndpointCheck{
			Kind:       "metadata",
			Address:    broker.Address,
			Reachable:  result.Reachable,
			DurationMs: result.Duration.Milliseconds(),
			Error:      result.Error,
		})
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

	seen := map[string]struct{}{}
	for _, service := range composeutil.KafkaServices(compose) {
		voters, err := composeutil.ParseVoters(service.Environment["KAFKA_CFG_CONTROLLER_QUORUM_VOTERS"])
		if err != nil {
			continue
		}
		for _, address := range voters {
			if _, ok := seen[address]; ok {
				continue
			}
			seen[address] = struct{}{}
			result := tcp.Dial(ctx, address, env.TCPTimeout)
			network.ControllerChecks = append(network.ControllerChecks, snapshot.EndpointCheck{
				Kind:       "controller",
				Address:    address,
				Reachable:  result.Reachable,
				DurationMs: result.Duration.Milliseconds(),
				Error:      result.Error,
			})
		}
	}
	return network
}
