package snapshot

type Bundle struct {
	Compose *ComposeSnapshot `json:"compose,omitempty"`
	Host    *HostSnapshot    `json:"host,omitempty"`
	Docker  *DockerSnapshot  `json:"docker,omitempty"`
	Network *NetworkSnapshot `json:"network,omitempty"`
	Kafka   *KafkaSnapshot   `json:"kafka,omitempty"`
	Topic   *TopicSnapshot   `json:"topic,omitempty"`
	Probe   *ProbeSnapshot   `json:"probe,omitempty"`
	Logs    *LogSnapshot     `json:"logs,omitempty"`
}
