package kafka

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/IBM/sarama"
)

type Metadata struct {
	ClusterID         string
	Brokers           []BrokerMetadata
	ControllerID      *int32
	ControllerAddress string
	Topics            []TopicMetadata
}

type BrokerMetadata struct {
	ID      int32
	Address string
}

type TopicMetadata struct {
	Name       string
	Partitions []PartitionMetadata
}

type PartitionMetadata struct {
	ID       int32
	LeaderID *int32
	Replicas []int32
	ISR      []int32
}

type ProbeMessage struct {
	MessageID  string `json:"message_id"`
	Mode       string `json:"mode"`
	SourceHost string `json:"source_host"`
	SentAtUnix int64  `json:"sent_at_unix"`
	Padding    string `json:"padding,omitempty"`
}

type ProduceResult struct {
	Partition  int32
	Offset     int64
	DurationMs int64
}

type ConsumeResult struct {
	DurationMs int64
	MessageID  string
}

func FetchMetadata(brokers []string, timeout time.Duration) (*Metadata, error) {
	cfg := newConfig(timeout)

	client, err := sarama.NewClient(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}
	defer client.Close()

	out := &Metadata{}
	for _, broker := range client.Brokers() {
		out.Brokers = append(out.Brokers, BrokerMetadata{
			ID:      broker.ID(),
			Address: broker.Addr(),
		})
	}
	if controller, err := client.Controller(); err == nil && controller != nil {
		id := controller.ID()
		out.ControllerID = &id
		out.ControllerAddress = controller.Addr()
	}

	topics, err := client.Topics()
	if err != nil {
		return nil, fmt.Errorf("list topics: %w", err)
	}
	for _, topic := range topics {
		partitions, err := client.Partitions(topic)
		if err != nil {
			continue
		}
		topicMeta := TopicMetadata{Name: topic}
		for _, partition := range partitions {
			p := PartitionMetadata{ID: partition}
			if leader, err := client.Leader(topic, partition); err == nil && leader != nil {
				id := leader.ID()
				p.LeaderID = &id
			}
			if replicas, err := client.Replicas(topic, partition); err == nil {
				p.Replicas = append([]int32(nil), replicas...)
			}
			if isr, err := client.InSyncReplicas(topic, partition); err == nil {
				p.ISR = append([]int32(nil), isr...)
			}
			topicMeta.Partitions = append(topicMeta.Partitions, p)
		}
		out.Topics = append(out.Topics, topicMeta)
	}

	return out, nil
}

func ProduceProbeMessage(brokers []string, timeout time.Duration, topic string, payload []byte) (*ProduceResult, error) {
	cfg := newConfig(timeout)
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true

	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("create sync producer: %w", err)
	}
	defer producer.Close()

	startedAt := time.Now()
	partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(payload),
	})
	if err != nil {
		return nil, fmt.Errorf("send probe message: %w", err)
	}

	return &ProduceResult{
		Partition:  partition,
		Offset:     offset,
		DurationMs: time.Since(startedAt).Milliseconds(),
	}, nil
}

func ConsumeProbeMessage(brokers []string, timeout time.Duration, topic string, partition int32, offset int64) (*ConsumeResult, error) {
	cfg := newConfig(timeout)
	consumer, err := sarama.NewConsumer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("create consumer: %w", err)
	}
	defer consumer.Close()

	pc, err := consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		return nil, fmt.Errorf("consume partition: %w", err)
	}
	defer pc.Close()

	startedAt := time.Now()
	select {
	case msg := <-pc.Messages():
		var payload ProbeMessage
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			return nil, fmt.Errorf("decode consumed probe message: %w", err)
		}
		return &ConsumeResult{
			DurationMs: time.Since(startedAt).Milliseconds(),
			MessageID:  payload.MessageID,
		}, nil
	case err := <-pc.Errors():
		return nil, fmt.Errorf("consume probe message: %w", err)
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for probe message")
	}
}

func CommitProbeOffset(brokers []string, timeout time.Duration, groupID string, topic string, partition int32, nextOffset int64) (int64, error) {
	cfg := newConfig(timeout)
	client, err := sarama.NewClient(brokers, cfg)
	if err != nil {
		return 0, fmt.Errorf("create client for offset manager: %w", err)
	}
	defer client.Close()

	startedAt := time.Now()
	manager, err := sarama.NewOffsetManagerFromClient(groupID, client)
	if err != nil {
		return 0, fmt.Errorf("create offset manager: %w", err)
	}
	defer manager.Close()

	partitionManager, err := manager.ManagePartition(topic, partition)
	if err != nil {
		return 0, fmt.Errorf("manage partition offset: %w", err)
	}
	defer partitionManager.Close()

	partitionManager.MarkOffset(nextOffset, "kdoctor probe")
	manager.Commit()
	return time.Since(startedAt).Milliseconds(), nil
}

func BuildProbePayload(messageID, mode string, size int) ([]byte, error) {
	host, _ := os.Hostname()
	payload := ProbeMessage{
		MessageID:  messageID,
		Mode:       mode,
		SourceHost: host,
		SentAtUnix: time.Now().UnixNano(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal probe payload: %w", err)
	}
	if size <= len(data) {
		return data, nil
	}

	payload.Padding = padWithX(size - len(data))
	data, err = json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal padded probe payload: %w", err)
	}
	return data, nil
}

func newConfig(timeout time.Duration) *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_0_0_0
	cfg.Net.DialTimeout = timeout
	cfg.Net.ReadTimeout = timeout
	cfg.Net.WriteTimeout = timeout
	cfg.Metadata.Timeout = timeout
	cfg.Consumer.Return.Errors = true
	return cfg
}

func padWithX(count int) string {
	if count <= 0 {
		return ""
	}
	buf := make([]byte, count)
	for i := range buf {
		buf[i] = 'x'
	}
	return string(buf)
}
