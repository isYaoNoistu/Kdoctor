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

	probe := &snapshot.ProbeSnapshot{
		Topic:            env.ProbeTopic,
		GroupID:          fmt.Sprintf("%s-%d", env.ProbeGroupPrefix, time.Now().UnixNano()),
		MessageID:        fmt.Sprintf("probe-%d", time.Now().UnixNano()),
		BootstrapAddress: brokers[0],
	}

	startedAt := time.Now()
	select {
	case <-ctx.Done():
		probe.FailureStage = "context"
		probe.Error = ctx.Err().Error()
		return probe
	default:
	}

	bootstrapStartedAt := time.Now()
	probe.BootstrapOK = true
	probe.BootstrapDurationMs = time.Since(bootstrapStartedAt).Milliseconds()

	metadataStartedAt := time.Now()
	if _, err := kafkatransport.FetchMetadata(brokers, env.ProbeTimeout); err != nil {
		probe.MetadataDurationMs = time.Since(metadataStartedAt).Milliseconds()
		probe.FailureStage = "metadata"
		probe.Error = err.Error()
		return probe
	}
	probe.MetadataOK = true
	probe.MetadataDurationMs = time.Since(metadataStartedAt).Milliseconds()

	produced, err := Produce(env, brokers, probe.MessageID)
	if err != nil {
		probe.FailureStage = "produce"
		probe.Error = err.Error()
		return probe
	}
	probe.ProduceOK = true
	probe.ProduceDurationMs = produced.DurationMs
	probe.ProducedPartition = produced.Partition
	probe.ProducedOffset = produced.Offset

	consumed, err := Consume(env, brokers, produced.Partition, produced.Offset)
	if err != nil {
		probe.FailureStage = "consume"
		probe.Error = err.Error()
		return probe
	}
	if consumed.MessageID != probe.MessageID {
		probe.FailureStage = "consume"
		probe.Error = fmt.Sprintf("message mismatch: expected %s got %s", probe.MessageID, consumed.MessageID)
		return probe
	}
	probe.ConsumeOK = true
	probe.ConsumeDurationMs = consumed.DurationMs

	commitDurationMs, err := Commit(env, brokers, probe.GroupID, produced.Partition, produced.Offset+1)
	if err != nil {
		probe.FailureStage = "commit"
		probe.Error = err.Error()
		return probe
	}
	probe.CommitOK = true
	probe.CommitDurationMs = commitDurationMs
	probe.EndToEndDurationMs = time.Since(startedAt).Milliseconds()
	return probe
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
