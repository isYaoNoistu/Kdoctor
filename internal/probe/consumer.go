package probe

import (
	"fmt"

	"kdoctor/internal/config"
	kafkatransport "kdoctor/internal/transport/kafka"
)

func Consume(env *config.Runtime, brokers []string, partition int32, offset int64) (*kafkatransport.ConsumeResult, error) {
	result, err := kafkatransport.ConsumeProbeMessage(brokers, env.ProbeTimeout, env.ProbeTopic, partition, offset)
	if err != nil {
		return nil, fmt.Errorf("consume probe: %w", err)
	}
	return result, nil
}

func Commit(env *config.Runtime, brokers []string, groupID string, partition int32, nextOffset int64) (int64, error) {
	durationMs, err := kafkatransport.CommitProbeOffset(brokers, env.ProbeTimeout, groupID, env.ProbeTopic, partition, nextOffset)
	if err != nil {
		return 0, fmt.Errorf("commit probe offset: %w", err)
	}
	return durationMs, nil
}
