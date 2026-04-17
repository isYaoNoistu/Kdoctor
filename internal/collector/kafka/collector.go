package kafka

import (
	"context"
	"errors"

	"kdoctor/internal/config"
	"kdoctor/internal/snapshot"
	kafkatransport "kdoctor/internal/transport/kafka"
)

type Collector struct{}

func (Collector) Collect(_ context.Context, env *config.Runtime, network *snapshot.NetworkSnapshot) (*snapshot.KafkaSnapshot, *snapshot.TopicSnapshot, error) {
	var brokers []string
	for _, check := range network.BootstrapChecks {
		if check.Reachable {
			brokers = append(brokers, check.Address)
		}
	}
	if len(brokers) == 0 {
		brokers = append(brokers, env.BootstrapExternal...)
	}
	if len(brokers) == 0 {
		brokers = append(brokers, env.BootstrapInternal...)
	}
	if len(brokers) == 0 {
		return nil, nil, errors.New("no bootstrap brokers configured")
	}

	meta, err := kafkatransport.FetchMetadata(brokers, env.MetadataTimeout)
	if err != nil {
		return nil, nil, err
	}

	kafkaSnap := &snapshot.KafkaSnapshot{
		ClusterID:           meta.ClusterID,
		ControllerID:        meta.ControllerID,
		ControllerAddress:   meta.ControllerAddress,
		ExpectedBrokerCount: env.SelectedProfile.BrokerCount,
	}
	for _, broker := range meta.Brokers {
		kafkaSnap.Brokers = append(kafkaSnap.Brokers, snapshot.BrokerSnapshot{
			ID:      broker.ID,
			Address: broker.Address,
		})
	}

	topicSnap := &snapshot.TopicSnapshot{}
	for _, topic := range meta.Topics {
		topicInfo := snapshot.TopicInfo{Name: topic.Name}
		for _, partition := range topic.Partitions {
			topicInfo.Partitions = append(topicInfo.Partitions, snapshot.PartitionInfo{
				ID:       partition.ID,
				LeaderID: partition.LeaderID,
				Replicas: append([]int32(nil), partition.Replicas...),
				ISR:      append([]int32(nil), partition.ISR...),
			})
		}
		topicSnap.Topics = append(topicSnap.Topics, topicInfo)
	}

	return kafkaSnap, topicSnap, nil
}
