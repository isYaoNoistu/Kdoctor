package snapshot

type ProbeSnapshot struct {
	Skipped             bool   `json:"skipped"`
	Reason              string `json:"reason,omitempty"`
	Topic               string `json:"topic,omitempty"`
	GroupID             string `json:"group_id,omitempty"`
	MessageID           string `json:"message_id,omitempty"`
	BootstrapOK         bool   `json:"bootstrap_ok"`
	MetadataOK          bool   `json:"metadata_ok"`
	ProduceOK           bool   `json:"produce_ok"`
	ConsumeOK           bool   `json:"consume_ok"`
	CommitOK            bool   `json:"commit_ok"`
	BootstrapAddress    string `json:"bootstrap_address,omitempty"`
	ProducedPartition   int32  `json:"produced_partition,omitempty"`
	ProducedOffset      int64  `json:"produced_offset,omitempty"`
	BootstrapDurationMs int64  `json:"bootstrap_duration_ms,omitempty"`
	MetadataDurationMs  int64  `json:"metadata_duration_ms,omitempty"`
	ProduceDurationMs   int64  `json:"produce_duration_ms,omitempty"`
	ConsumeDurationMs   int64  `json:"consume_duration_ms,omitempty"`
	CommitDurationMs    int64  `json:"commit_duration_ms,omitempty"`
	EndToEndDurationMs  int64  `json:"end_to_end_duration_ms,omitempty"`
	FailureStage        string `json:"failure_stage,omitempty"`
	Error               string `json:"error,omitempty"`
}
