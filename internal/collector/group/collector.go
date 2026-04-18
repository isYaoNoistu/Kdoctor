package group

import (
	"context"
	"strings"

	"kdoctor/internal/config"
	"kdoctor/internal/snapshot"
	kafkatransport "kdoctor/internal/transport/kafka"
)

type Collector struct{}

func (Collector) Collect(_ context.Context, env *config.Runtime, network *snapshot.NetworkSnapshot) *snapshot.GroupSnapshot {
	if env == nil || len(env.SelectedProfile.GroupProbeTargets) == 0 {
		return nil
	}

	brokers := selectBrokers(env, network)
	out := &snapshot.GroupSnapshot{Collected: true}
	if len(brokers) == 0 {
		out.Errors = append(out.Errors, "no bootstrap brokers available for consumer group inspection")
		return out
	}

	targets := make([]kafkatransport.ConsumerGroupLagTarget, 0, len(env.SelectedProfile.GroupProbeTargets))
	for _, target := range env.SelectedProfile.GroupProbeTargets {
		groupID := strings.TrimSpace(target.GroupID)
		if groupID == "" {
			groupID = strings.TrimSpace(target.Name)
		}
		if groupID == "" || strings.TrimSpace(target.Topic) == "" {
			continue
		}
		targets = append(targets, kafkatransport.ConsumerGroupLagTarget{
			Name:    target.Name,
			GroupID: groupID,
			Topic:   target.Topic,
		})
	}
	if len(targets) == 0 {
		out.Errors = append(out.Errors, "no valid consumer group targets were configured")
		return out
	}

	results, err := kafkatransport.FetchConsumerGroupLag(brokers, env.AdminAPITimeout, targets)
	if err != nil {
		out.Errors = append(out.Errors, err.Error())
		return out
	}

	for _, result := range results {
		item := snapshot.GroupLagSnapshot{
			Name:            result.Name,
			GroupID:         result.GroupID,
			Topic:           result.Topic,
			State:           result.State,
			Coordinator:     result.Coordinator,
			MemberCount:     result.MemberCount,
			TotalLag:        result.TotalLag,
			MaxPartitionLag: result.MaxPartitionLag,
			MaxLagPartition: result.MaxLagPartition,
			MissingOffsets:  result.MissingOffsets,
			Error:           result.Error,
		}
		for _, partition := range result.Partitions {
			item.Partitions = append(item.Partitions, snapshot.GroupPartitionLagInfo{
				Partition:          partition.Partition,
				CommittedOffset:    partition.CommittedOffset,
				EndOffset:          partition.EndOffset,
				Lag:                partition.Lag,
				HasCommittedOffset: partition.HasCommittedOffset,
			})
		}
		out.Targets = append(out.Targets, item)
	}

	out.Available = len(out.Targets) > 0
	return out
}

func selectBrokers(env *config.Runtime, network *snapshot.NetworkSnapshot) []string {
	brokers := []string{}
	if network != nil {
		for _, check := range network.BootstrapChecks {
			if check.Reachable {
				brokers = append(brokers, check.Address)
			}
		}
	}
	if len(brokers) == 0 {
		brokers = append(brokers, env.BootstrapExternal...)
	}
	if len(brokers) == 0 {
		brokers = append(brokers, env.BootstrapInternal...)
	}
	return dedupeStrings(brokers)
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
