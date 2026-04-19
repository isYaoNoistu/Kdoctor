package probe

import (
	"context"
	"fmt"
	"time"

	"kdoctor/internal/config"
	"kdoctor/internal/snapshot"
	kafkatransport "kdoctor/internal/transport/kafka"
)

func Run(ctx context.Context, env *config.Runtime) *snapshot.ProbeSnapshot {
	if !shouldRun(env.Mode, env.Config.Probe.Enabled) {
		return &snapshot.ProbeSnapshot{
			Skipped: true,
			Reason:  "probe disabled for current mode",
		}
	}
	if env.ProbeTopic == "" {
		return &snapshot.ProbeSnapshot{
			Skipped: true,
			Reason:  "probe topic is not configured",
		}
	}

	brokers := availableBrokers(env)
	if len(brokers) == 0 {
		return &snapshot.ProbeSnapshot{
			Skipped: true,
			Reason:  "no probe brokers are available",
		}
	}

	baseMessageID := fmt.Sprintf("probe-%d", time.Now().UnixNano())
	probe := &snapshot.ProbeSnapshot{
		Topic:                env.ProbeTopic,
		GroupID:              fmt.Sprintf("%s-%d", env.ProbeGroupPrefix, time.Now().UnixNano()),
		MessageID:            baseMessageID,
		BootstrapAddress:     brokers[0],
		ProducedMessageCount: max(1, env.ProbeProduceCount),
		ExecutedStage:        snapshot.ProbeStageBootstrap,
	}

	startedAt := time.Now()
	defer cleanupProbeTopic(env, brokers, probe)
	defer finalizeProbe(probe, startedAt)

	if !checkContext(ctx, probe) {
		return probe
	}

	bootstrapStartedAt := time.Now()
	probe.BootstrapOK = true
	probe.BootstrapDurationMs = time.Since(bootstrapStartedAt).Milliseconds()

	metadataStartedAt := time.Now()
	probe.MetadataExecuted = true
	meta, err := kafkatransport.FetchMetadata(brokers, env.ProbeTimeout)
	probe.MetadataDurationMs = time.Since(metadataStartedAt).Milliseconds()
	if err != nil {
		failProbe(probe, snapshot.ProbeStageMetadata, err)
		return probe
	}
	probe.MetadataOK = true
	probe.ExecutedStage = snapshot.ProbeStageMetadata

	ready, created, reason, err := ensureTopicReady(env, brokers, meta)
	probe.TopicReady = ready
	probe.TopicCreated = created
	probe.TopicReadyReason = reason
	if err != nil {
		failProbe(probe, snapshot.ProbeStageTopicReady, err)
		return probe
	}
	probe.ExecutedStage = snapshot.ProbeStageTopicReady

	if !checkContext(ctx, probe) {
		return probe
	}

	probe.ProduceExecuted = true
	produceStartedAt := time.Now()
	var producedPartition int32
	var producedOffset int64
	for i := 0; i < max(1, env.ProbeProduceCount); i++ {
		currentMessageID := baseMessageID
		if env.ProbeProduceCount > 1 {
			currentMessageID = fmt.Sprintf("%s-%d", baseMessageID, i+1)
		}

		produced, produceErr := Produce(env, brokers, currentMessageID)
		if produceErr != nil {
			probe.MessageID = currentMessageID
			probe.ProducedMessageCount = i
			probe.ProduceDurationMs = time.Since(produceStartedAt).Milliseconds()
			failProbe(probe, snapshot.ProbeStageProduce, produceErr)
			return probe
		}

		probe.MessageID = currentMessageID
		probe.ProducedMessageCount = i + 1
		producedPartition = produced.Partition
		producedOffset = produced.Offset
		probe.ProducedPartition = producedPartition
		probe.ProducedOffset = producedOffset
	}
	probe.ProduceOK = true
	probe.ProduceDurationMs = time.Since(produceStartedAt).Milliseconds()
	probe.ExecutedStage = snapshot.ProbeStageProduce

	if !checkContext(ctx, probe) {
		return probe
	}

	probe.ConsumeExecuted = true
	consumed, err := Consume(env, brokers, producedPartition, producedOffset)
	if err != nil {
		failProbe(probe, snapshot.ProbeStageConsume, err)
		return probe
	}
	if consumed.MessageID != probe.MessageID {
		failProbe(probe, snapshot.ProbeStageConsume, fmt.Errorf("message mismatch: expected %s got %s", probe.MessageID, consumed.MessageID))
		return probe
	}
	probe.ConsumeOK = true
	probe.ConsumeDurationMs = consumed.DurationMs
	probe.ExecutedStage = snapshot.ProbeStageConsume

	if !checkContext(ctx, probe) {
		return probe
	}

	probe.CommitExecuted = true
	commitDurationMs, err := Commit(env, brokers, probe.GroupID, producedPartition, producedOffset+1)
	if err != nil {
		failProbe(probe, snapshot.ProbeStageCommit, err)
		return probe
	}
	probe.CommitOK = true
	probe.CommitDurationMs = commitDurationMs
	probe.ExecutedStage = snapshot.ProbeStageCommit

	probe.ExecutedStage = snapshot.ProbeStageComplete
	return probe
}

func ensureTopicReady(env *config.Runtime, brokers []string, meta *kafkatransport.Metadata) (bool, bool, string, error) {
	if kafkatransport.TopicExists(meta, env.ProbeTopic) {
		if err := waitForTopicLeaderReady(brokers, env.ProbeTimeout, env.ProbeTopic); err != nil {
			return false, false, "probe topic exists but leader is not ready yet", err
		}
		return true, false, "probe topic already exists", nil
	}

	replicationFactor := desiredProbeReplicationFactor(env, meta)
	created, err := kafkatransport.EnsureProbeTopic(brokers, env.ProbeTimeout, env.ProbeTopic, 1, replicationFactor)
	if err != nil {
		return false, created, "probe topic could not be prepared", fmt.Errorf("ensure probe topic: %w", err)
	}
	if created {
		if err := waitForTopicLeaderReady(brokers, env.ProbeTimeout, env.ProbeTopic); err != nil {
			return false, true, "probe topic created but leader is not ready yet", err
		}
		return true, true, "probe topic created for this run", nil
	}
	if err := waitForTopicLeaderReady(brokers, env.ProbeTimeout, env.ProbeTopic); err != nil {
		return false, false, "probe topic became available but leader is not ready yet", err
	}
	return true, false, "probe topic became available during readiness check", nil
}

func waitForTopicLeaderReady(brokers []string, timeout time.Duration, topic string) error {
	deadline := time.Now().Add(timeout)
	for {
		meta, err := kafkatransport.FetchMetadata(brokers, timeout)
		if err == nil && topicHasReadyLeader(meta, topic) {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("probe topic %q exists in metadata but leader was not ready before timeout", topic)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func topicHasReadyLeader(meta *kafkatransport.Metadata, topic string) bool {
	if meta == nil {
		return false
	}
	for _, item := range meta.Topics {
		if item.Name != topic {
			continue
		}
		if len(item.Partitions) == 0 {
			return false
		}
		for _, partition := range item.Partitions {
			if partition.LeaderID == nil {
				return false
			}
		}
		return true
	}
	return false
}

func cleanupProbeTopic(env *config.Runtime, brokers []string, probe *snapshot.ProbeSnapshot) {
	if probe == nil || !env.Config.Probe.Cleanup || !probe.TopicCreated {
		return
	}

	probe.CleanupAttempted = true
	if err := kafkatransport.DeleteProbeTopic(brokers, env.ProbeTimeout, env.ProbeTopic); err != nil {
		probe.CleanupError = err.Error()
		return
	}
	probe.CleanupOK = true
}

func desiredProbeReplicationFactor(env *config.Runtime, meta *kafkatransport.Metadata) int16 {
	available := 1
	if meta != nil && len(meta.Brokers) > 0 {
		available = len(meta.Brokers)
	}
	if env.SelectedProfile.ExpectedReplicationFactor > 0 {
		if env.SelectedProfile.ExpectedReplicationFactor < available {
			return int16(env.SelectedProfile.ExpectedReplicationFactor)
		}
		return int16(available)
	}
	return int16(available)
}

func checkContext(ctx context.Context, probe *snapshot.ProbeSnapshot) bool {
	select {
	case <-ctx.Done():
		failProbe(probe, snapshot.ProbeStageContext, ctx.Err())
		return false
	default:
		return true
	}
}

func failProbe(probe *snapshot.ProbeSnapshot, stage string, err error) {
	if probe == nil {
		return
	}
	probe.FailureStage = stage
	if err != nil {
		probe.Error = err.Error()
	}
	if probe.DownstreamSkippedHint == "" {
		switch stage {
		case snapshot.ProbeStageMetadata:
			probe.DownstreamSkippedHint = "metadata stage failed; downstream probe stages were skipped"
		case snapshot.ProbeStageTopicReady:
			probe.DownstreamSkippedHint = "probe topic was not ready; downstream probe stages were skipped"
		case snapshot.ProbeStageProduce:
			probe.DownstreamSkippedHint = "produce stage failed; downstream probe stages were skipped"
		case snapshot.ProbeStageConsume:
			probe.DownstreamSkippedHint = "consume stage failed; downstream probe stages were skipped"
		case snapshot.ProbeStageCommit:
			probe.DownstreamSkippedHint = "commit stage failed; end-to-end probe ended at commit stage"
		case snapshot.ProbeStageContext:
			probe.DownstreamSkippedHint = "execution context ended before the probe finished"
		}
	}
}

func finalizeProbe(probe *snapshot.ProbeSnapshot, startedAt time.Time) {
	if probe == nil {
		return
	}
	probe.EndToEndDurationMs = time.Since(startedAt).Milliseconds()
}

func shouldRun(mode string, enabled bool) bool {
	if !enabled {
		return false
	}
	switch mode {
	case "probe", "full", "incident":
		return true
	default:
		return false
	}
}

func availableBrokers(env *config.Runtime) []string {
	if len(env.BootstrapExternal) > 0 {
		return append([]string(nil), env.BootstrapExternal...)
	}
	return append([]string(nil), env.BootstrapInternal...)
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
