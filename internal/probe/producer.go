package probe

import (
	"fmt"

	"kdoctor/internal/config"
	kafkatransport "kdoctor/internal/transport/kafka"
)

func Produce(env *config.Runtime, brokers []string, messageID string) (*kafkatransport.ProduceResult, error) {
	payload, err := kafkatransport.BuildProbePayload(messageID, env.Mode, env.ProbeMessageBytes)
	if err != nil {
		return nil, err
	}
	result, err := kafkatransport.ProduceProbeMessage(brokers, env.ProbeTimeout, env.ProbeTopic, payload)
	if err != nil {
		return nil, fmt.Errorf("produce probe: %w", err)
	}
	return result, nil
}
