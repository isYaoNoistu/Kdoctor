package snapshot

type TopicSnapshot struct {
	Topics []TopicInfo `json:"topics,omitempty"`
}

type TopicInfo struct {
	Name       string          `json:"name"`
	Partitions []PartitionInfo `json:"partitions,omitempty"`
}

type PartitionInfo struct {
	ID       int32   `json:"id"`
	LeaderID *int32  `json:"leader_id,omitempty"`
	Replicas []int32 `json:"replicas,omitempty"`
	ISR      []int32 `json:"isr,omitempty"`
}
