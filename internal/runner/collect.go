package runner

import (
	"context"
	"time"

	clientchecks "kdoctor/internal/checks/client"
	consumerchecks "kdoctor/internal/checks/consumer"
	dockerchecks "kdoctor/internal/checks/docker"
	hostchecks "kdoctor/internal/checks/host"
	kafkachecks "kdoctor/internal/checks/kafka"
	kraftchecks "kdoctor/internal/checks/kraft"
	lintchecks "kdoctor/internal/checks/lint"
	logchecks "kdoctor/internal/checks/logs"
	metricschecks "kdoctor/internal/checks/metrics"
	networkchecks "kdoctor/internal/checks/network"
	producerchecks "kdoctor/internal/checks/producer"
	securitychecks "kdoctor/internal/checks/security"
	storagechecks "kdoctor/internal/checks/storage"
	topicchecks "kdoctor/internal/checks/topic"
	transactionchecks "kdoctor/internal/checks/transaction"
	upgradechecks "kdoctor/internal/checks/upgrade"
	composecollector "kdoctor/internal/collector/compose"
	dockercollector "kdoctor/internal/collector/docker"
	groupcollector "kdoctor/internal/collector/group"
	hostcollector "kdoctor/internal/collector/host"
	kafkacollector "kdoctor/internal/collector/kafka"
	logcollector "kdoctor/internal/collector/logs"
	metricscollector "kdoctor/internal/collector/metrics"
	networkcollector "kdoctor/internal/collector/network"
	"kdoctor/internal/config"
	"kdoctor/internal/probe"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type checker interface {
	ID() string
	Name() string
	Module() string
	Run(ctx context.Context, snap *snapshot.Bundle) model.CheckResult
}

func CollectAndCheck(ctx context.Context, env *config.Runtime) (*snapshot.Bundle, []model.CheckResult, []string) {
	bundle := &snapshot.Bundle{}
	errs := []string{}

	var composeSnap *snapshot.ComposeSnapshot
	var networkSnap *snapshot.NetworkSnapshot

	errs = append(errs, runTasks(
		ctx,
		taskSpec{
			Name:    "compose_snapshot",
			Timeout: minDuration(env.MetadataTimeout, 5*time.Second),
			Soft:    true,
			Run: func(taskCtx context.Context) error {
				var err error
				composeSnap, err = composecollector.Collector{}.Collect(taskCtx, env)
				return err
			},
		},
		taskSpec{
			Name:    "network_base",
			Timeout: minDuration(env.MetadataTimeout, 10*time.Second),
			Soft:    true,
			Run: func(taskCtx context.Context) error {
				networkSnap = networkcollector.Collector{}.CollectBase(taskCtx, env)
				return nil
			},
		},
	)...)

	bundle.Compose = composeSnap
	bundle.Network = networkcollector.Collector{}.CollectComposeControllers(ctx, env, networkSnap, composeSnap)
	bundle.Metrics = metricscollector.Collector{}.Collect(ctx, env, composeSnap)

	kafkaSnap, topicSnap, kafkaErr := kafkacollector.Collector{}.Collect(ctx, env, networkSnap)
	if kafkaErr != nil {
		errs = append(errs, kafkaErr.Error())
	} else {
		bundle.Kafka = kafkaSnap
		bundle.Topic = topicSnap
		bundle.Network = networkcollector.Collector{}.CollectMetadata(ctx, env, networkSnap, kafkaSnap.Brokers)
		bundle.Group = groupcollector.Collector{}.Collect(ctx, env, bundle.Network)
	}
	bundle.Docker = dockercollector.Collector{}.Collect(ctx, env, composeSnap)
	bundle.Host = hostcollector.Collector{}.Collect(ctx, env, composeSnap, bundle.Docker)
	bundle.Logs = logcollector.Collector{}.Collect(ctx, env, composeSnap, bundle.Docker)
	bundle.Probe = probe.Run(ctx, env)
	if bundle.Probe != nil && bundle.Probe.ShouldRefreshKafkaSnapshot() {
		refreshedKafka, refreshedTopic, refreshErr := kafkacollector.Collector{}.Collect(ctx, env, bundle.Network)
		if refreshErr != nil {
			errs = append(errs, refreshErr.Error())
		} else {
			bundle.Kafka = refreshedKafka
			bundle.Topic = refreshedTopic
			bundle.Network = networkcollector.Collector{}.CollectMetadata(ctx, env, bundle.Network, refreshedKafka.Brokers)
		}
	}

	checks := runChecks(ctx, env, bundle)
	return bundle, checks, errs
}

func minDuration(left time.Duration, right time.Duration) time.Duration {
	if left <= 0 {
		return right
	}
	if right <= 0 {
		return left
	}
	if left < right {
		return left
	}
	return right
}

func runChecks(ctx context.Context, env *config.Runtime, bundle *snapshot.Bundle) []model.CheckResult {
	checkers := []checker{
		networkchecks.BootstrapChecker{},
		networkchecks.ListenerChecker{},
		networkchecks.MetadataChecker{},
		networkchecks.RouteMismatchChecker{},
		networkchecks.AdvertisedPrivateChecker{},
		networkchecks.BootstrapLBChecker{},
		networkchecks.DNSChecker{},
		networkchecks.DNSDriftChecker{},
		networkchecks.ProtocolMismatchChecker{},
		kafkachecks.ClusterChecker{},
		kafkachecks.RegistrationChecker{},
		kafkachecks.RegistrationIntegrityChecker{},
		kafkachecks.EndpointChecker{},
		kafkachecks.ReturnedRouteChecker{},
		kafkachecks.BrokerIdentityChecker{},
		kafkachecks.InternalTopicsChecker{},
		kafkachecks.MetadataLatencyChecker{},
		kafkachecks.TopologyMismatchChecker{},
		kraftchecks.ConfigChecker{},
		kraftchecks.ControllerChecker{},
		kraftchecks.QuorumChecker{},
		kraftchecks.MajorityChecker{},
		kraftchecks.EndpointConfigChecker{},
		kraftchecks.EpochChecker{},
		kraftchecks.UnknownVoterChecker{},
		kraftchecks.FinalizationChecker{},
		topicchecks.LeaderChecker{},
		topicchecks.ReplicaHealthChecker{},
		topicchecks.ISRChecker{MinISR: env.SelectedProfile.ExpectedMinISR},
		topicchecks.UnderReplicatedChecker{WarnCount: env.Config.Thresholds.URPWarn},
		topicchecks.UnderMinISRChecker{MinISR: env.SelectedProfile.ExpectedMinISR},
		topicchecks.OfflineReplicaChecker{},
		topicchecks.LeaderSkewChecker{WarnPct: env.Config.Thresholds.LeaderSkewWarnPct},
		topicchecks.ReplicaLagChecker{Warn: env.Config.Thresholds.ReplicaLagWarn},
		topicchecks.PlanningChecker{ExpectedBrokerCount: env.SelectedProfile.BrokerCount},
		lintchecks.ComposeChecker{},
		lintchecks.NodeIDChecker{},
		lintchecks.ClusterIDChecker{},
		lintchecks.ProcessRolesChecker{RequireBroker: true, RequireController: true},
		lintchecks.QuorumVotersChecker{},
		lintchecks.ListenersChecker{},
		lintchecks.AdvertisedViewChecker{ExecutionView: env.SelectedProfile.ExecutionView},
		lintchecks.ControllerListenerChecker{},
		lintchecks.BrokerIdentityChecker{},
		lintchecks.TopologyChecker{
			ExpectedBrokerCount: env.SelectedProfile.BrokerCount,
			ExpectedControllers: len(env.SelectedProfile.ControllerEndpoints),
		},
		lintchecks.TopicPlanningChecker{},
		lintchecks.MetadataDirChecker{},
		lintchecks.InterBrokerListenerChecker{},
		lintchecks.ReplicationChecker{ExpectedBrokerCount: env.SelectedProfile.BrokerCount},
		securitychecks.ListenerChecker{
			ExecutionView: env.SelectedProfile.ExecutionView,
			SecurityMode:  env.SelectedProfile.SecurityMode,
		},
		securitychecks.SASLChecker{
			ExecutionView: env.SelectedProfile.ExecutionView,
			SecurityMode:  env.SelectedProfile.SecurityMode,
			SASLMechanism: env.SelectedProfile.SASLMechanism,
		},
		securitychecks.TLSChecker{
			ExecutionView:    env.SelectedProfile.ExecutionView,
			CertWarnDays:     env.Config.Thresholds.CertExpiryWarnDays,
			HandshakeTimeout: env.TCPTimeout,
		},
		securitychecks.AuthorizationChecker{},
		securitychecks.AuthorizerChecker{},
		storagechecks.CapacityChecker{
			DiskWarnPct:  env.Config.Thresholds.DiskWarnPct,
			DiskCritPct:  env.Config.Thresholds.DiskCritPct,
			InodeWarnPct: env.Config.Thresholds.InodeWarnPct,
		},
		storagechecks.OfflineLogDirChecker{},
		storagechecks.LayoutChecker{},
		storagechecks.PartialFailureChecker{},
		storagechecks.MountPlanningChecker{},
		storagechecks.TieredStorageChecker{},
		metricschecks.UnderReplicatedChecker{WarnCount: env.Config.Thresholds.URPWarn},
		metricschecks.MinISRChecker{UnderMinISRCrit: env.Config.Thresholds.UnderMinISRCrit},
		metricschecks.OfflineLogDirChecker{},
		metricschecks.ReplicaLagChecker{Warn: env.Config.Thresholds.ReplicaLagWarn},
		metricschecks.NetworkIdleChecker{Warn: env.Config.Thresholds.NetworkIdleWarn},
		metricschecks.RequestIdleChecker{Warn: env.Config.Thresholds.RequestIdleWarn},
		metricschecks.NetworkIdleMetricChecker{Warn: env.Config.Thresholds.NetworkIdleWarn},
		metricschecks.RequestIdleMetricChecker{Warn: env.Config.Thresholds.RequestIdleWarn},
		metricschecks.ProduceThrottleChecker{WarnMs: env.Config.Thresholds.ProduceThrottleWarnMs},
		metricschecks.FetchThrottleChecker{WarnMs: env.Config.Thresholds.FetchThrottleWarnMs},
		metricschecks.RequestQuotaChecker{},
		metricschecks.BackpressureChecker{RequestLatencyWarnMs: env.Config.Thresholds.RequestLatencyWarnMs},
		metricschecks.RequestPressureChecker{
			WarnLatencyMs: env.Config.Thresholds.RequestLatencyWarnMs,
			WarnPurgatory: env.Config.Thresholds.PurgatoryWarnCount,
		},
		metricschecks.HeapGCChecker{
			HeapWarnPct:   env.Config.Thresholds.HeapUsedWarnPct,
			GCPauseWarnMs: env.Config.Thresholds.GCPauseWarnMs,
		},
		producerchecks.AcksChecker{
			Acks:               env.SelectedProfile.Producer.Acks,
			ExpectedDurability: env.SelectedProfile.Producer.ExpectedDurability,
			MinISR:             env.SelectedProfile.ExpectedMinISR,
		},
		producerchecks.IdempotenceChecker{
			EnableIdempotence: env.SelectedProfile.Producer.EnableIdempotence,
			Retries:           env.SelectedProfile.Producer.Retries,
			MaxInFlight:       env.SelectedProfile.Producer.MaxInFlight,
		},
		producerchecks.TimeoutChecker{
			DeliveryTimeoutMs: env.SelectedProfile.Producer.DeliveryTimeoutMs,
			RequestTimeoutMs:  env.SelectedProfile.Producer.RequestTimeoutMs,
			LingerMs:          env.SelectedProfile.Producer.LingerMs,
		},
		producerchecks.MessageSizeChecker{
			ProbeMessageBytes: env.ProbeMessageBytes,
		},
		producerchecks.ThrottleChecker{
			WarnMs: env.Config.Thresholds.ProduceThrottleWarnMs,
		},
		producerchecks.TxTimeoutChecker{
			TransactionTimeoutMs: env.SelectedProfile.Producer.TransactionTimeoutMs,
		},
		clientchecks.MetadataChecker{},
		clientchecks.ProducerChecker{},
		clientchecks.ConsumerChecker{},
		clientchecks.CommitChecker{},
		clientchecks.EndToEndChecker{},
		consumerchecks.LagChecker{
			WarnLag: env.Config.Thresholds.ConsumerLagWarn,
			CritLag: env.Config.Thresholds.ConsumerLagCrit,
			Targets: env.SelectedProfile.GroupProbeTargets,
		},
		consumerchecks.RebalanceChecker{},
		consumerchecks.CoordinatorChecker{},
		consumerchecks.PollIntervalChecker{
			MaxPollIntervalMs: env.SelectedProfile.Consumer.MaxPollIntervalMs,
			SessionTimeoutMs:  env.SelectedProfile.Consumer.SessionTimeoutMs,
		},
		consumerchecks.HeartbeatChecker{
			SessionTimeoutMs:    env.SelectedProfile.Consumer.SessionTimeoutMs,
			HeartbeatIntervalMs: env.SelectedProfile.Consumer.HeartbeatIntervalMs,
		},
		consumerchecks.OffsetResetChecker{
			AutoOffsetReset: env.SelectedProfile.Consumer.AutoOffsetReset,
		},
		transactionchecks.TopicAbsenceChecker{
			TXProbeEnabled:  env.Config.Probe.TXProbeEnabled,
			TransactionalID: env.SelectedProfile.Producer.TransactionalID,
		},
		transactionchecks.RequiredTopicChecker{
			TXProbeEnabled:  env.Config.Probe.TXProbeEnabled,
			TransactionalID: env.SelectedProfile.Producer.TransactionalID,
		},
		transactionchecks.TimeoutChecker{
			TransactionTimeoutMs: env.SelectedProfile.Producer.TransactionTimeoutMs,
		},
		transactionchecks.IsolationChecker{
			IsolationLevel: env.SelectedProfile.Consumer.IsolationLevel,
			TXProbeEnabled: env.Config.Probe.TXProbeEnabled,
		},
		transactionchecks.OutcomeChecker{
			TXProbeEnabled:  env.Config.Probe.TXProbeEnabled,
			TransactionalID: env.SelectedProfile.Producer.TransactionalID,
		},
		upgradechecks.RollingVersionChecker{},
		upgradechecks.FeatureChecker{},
		upgradechecks.TieredStorageChecker{},
		hostchecks.DiskChecker{
			WarnPercent: env.Config.Thresholds.DiskWarnPct,
			CritPercent: env.Config.Thresholds.DiskCritPct,
		},
		hostchecks.CapacityChecker{
			DiskWarnPct:  env.Config.Thresholds.DiskWarnPct,
			DiskCritPct:  env.Config.Thresholds.DiskCritPct,
			InodeWarnPct: env.Config.Thresholds.InodeWarnPct,
		},
		hostchecks.FDChecker{
			WarnPct: env.Config.Host.FDWarnPct,
			CritPct: env.Config.Host.FDCritPct,
		},
		hostchecks.ClockChecker{
			WarnMs: env.Config.Host.ClockSkewWarnMs,
		},
		hostchecks.PortChecker{},
		hostchecks.ListenerDriftChecker{},
		hostchecks.MemoryChecker{
			WarnPct: env.Config.Thresholds.HeapUsedWarnPct,
		},
		dockerchecks.ExistenceChecker{},
		dockerchecks.RunningChecker{},
		dockerchecks.OOMChecker{},
		dockerchecks.RestartChecker{},
		dockerchecks.MemoryPlanningChecker{},
		dockerchecks.MountChecker{},
		dockerchecks.PersistenceChecker{},
		logchecks.SourcesChecker{},
		logchecks.FingerprintChecker{},
		logchecks.HitContextChecker{},
		logchecks.FreshnessChecker{},
		logchecks.StormChecker{},
		logchecks.CustomPatternChecker{},
		logchecks.ExplanationChecker{},
		logchecks.AggregateChecker{},
	}

	results := make([]model.CheckResult, 0, len(checkers))
	for _, c := range checkers {
		results = append(results, c.Run(ctx, bundle))
	}
	return results
}
