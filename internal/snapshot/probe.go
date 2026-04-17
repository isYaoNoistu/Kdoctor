package snapshot

const (
	ProbeStageBootstrap  = "bootstrap"
	ProbeStageMetadata   = "metadata"
	ProbeStageTopicReady = "topic_ready"
	ProbeStageProduce    = "produce"
	ProbeStageConsume    = "consume"
	ProbeStageCommit     = "commit"
	ProbeStageContext    = "context"
	ProbeStageComplete   = "complete"
)

type ProbeSnapshot struct {
	Skipped               bool   `json:"skipped"`
	Reason                string `json:"reason,omitempty"`
	Topic                 string `json:"topic,omitempty"`
	GroupID               string `json:"group_id,omitempty"`
	MessageID             string `json:"message_id,omitempty"`
	BootstrapOK           bool   `json:"bootstrap_ok"`
	TopicReady            bool   `json:"topic_ready"`
	TopicCreated          bool   `json:"topic_created"`
	MetadataOK            bool   `json:"metadata_ok"`
	ProduceOK             bool   `json:"produce_ok"`
	ConsumeOK             bool   `json:"consume_ok"`
	CommitOK              bool   `json:"commit_ok"`
	MetadataExecuted      bool   `json:"metadata_executed"`
	ProduceExecuted       bool   `json:"produce_executed"`
	ConsumeExecuted       bool   `json:"consume_executed"`
	CommitExecuted        bool   `json:"commit_executed"`
	CleanupAttempted      bool   `json:"cleanup_attempted"`
	CleanupOK             bool   `json:"cleanup_ok"`
	BootstrapAddress      string `json:"bootstrap_address,omitempty"`
	TopicReadyReason      string `json:"topic_ready_reason,omitempty"`
	ProducedPartition     int32  `json:"produced_partition,omitempty"`
	ProducedOffset        int64  `json:"produced_offset,omitempty"`
	ProducedMessageCount  int    `json:"produced_message_count,omitempty"`
	BootstrapDurationMs   int64  `json:"bootstrap_duration_ms,omitempty"`
	MetadataDurationMs    int64  `json:"metadata_duration_ms,omitempty"`
	ProduceDurationMs     int64  `json:"produce_duration_ms,omitempty"`
	ConsumeDurationMs     int64  `json:"consume_duration_ms,omitempty"`
	CommitDurationMs      int64  `json:"commit_duration_ms,omitempty"`
	EndToEndDurationMs    int64  `json:"end_to_end_duration_ms,omitempty"`
	FailureStage          string `json:"failure_stage,omitempty"`
	ExecutedStage         string `json:"executed_stage,omitempty"`
	DownstreamSkippedHint string `json:"downstream_skipped_hint,omitempty"`
	CleanupError          string `json:"cleanup_error,omitempty"`
	Error                 string `json:"error,omitempty"`
}

func (p *ProbeSnapshot) StageSkipReason(stage string) string {
	if p == nil || p.FailureStage == "" {
		return ""
	}

	switch stage {
	case ProbeStageProduce:
		switch p.FailureStage {
		case ProbeStageMetadata:
			return "metadata stage failed; produce stage was not executed"
		case ProbeStageTopicReady:
			return "probe topic was not ready; produce stage was not executed"
		case ProbeStageContext:
			return "execution context ended before produce stage"
		}
	case ProbeStageConsume:
		switch p.FailureStage {
		case ProbeStageMetadata:
			return "metadata stage failed; consume stage was not executed"
		case ProbeStageTopicReady:
			return "probe topic was not ready; consume stage was not executed"
		case ProbeStageProduce:
			return "produce stage failed; consume stage was not executed"
		case ProbeStageContext:
			return "execution context ended before consume stage"
		}
	case ProbeStageCommit:
		switch p.FailureStage {
		case ProbeStageMetadata:
			return "metadata stage failed; commit stage was not executed"
		case ProbeStageTopicReady:
			return "probe topic was not ready; commit stage was not executed"
		case ProbeStageProduce:
			return "produce stage failed; commit stage was not executed"
		case ProbeStageConsume:
			return "consume stage failed; commit stage was not executed"
		case ProbeStageContext:
			return "execution context ended before commit stage"
		}
	}

	if p.DownstreamSkippedHint != "" {
		return p.DownstreamSkippedHint
	}
	return ""
}

func (p *ProbeSnapshot) ShouldRefreshKafkaSnapshot() bool {
	if p == nil || p.Skipped {
		return false
	}
	return p.TopicCreated || p.CommitExecuted || p.CleanupAttempted
}
