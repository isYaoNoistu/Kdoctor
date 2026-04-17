package runner

import (
	"context"
	"time"

	clientchecks "kdoctor/internal/checks/client"
	dockerchecks "kdoctor/internal/checks/docker"
	hostchecks "kdoctor/internal/checks/host"
	kafkachecks "kdoctor/internal/checks/kafka"
	kraftchecks "kdoctor/internal/checks/kraft"
	lintchecks "kdoctor/internal/checks/lint"
	logchecks "kdoctor/internal/checks/logs"
	networkchecks "kdoctor/internal/checks/network"
	topicchecks "kdoctor/internal/checks/topic"
	composecollector "kdoctor/internal/collector/compose"
	dockercollector "kdoctor/internal/collector/docker"
	hostcollector "kdoctor/internal/collector/host"
	kafkacollector "kdoctor/internal/collector/kafka"
	logcollector "kdoctor/internal/collector/logs"
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

	kafkaSnap, topicSnap, kafkaErr := kafkacollector.Collector{}.Collect(ctx, env, networkSnap)
	if kafkaErr != nil {
		errs = append(errs, kafkaErr.Error())
	} else {
		bundle.Kafka = kafkaSnap
		bundle.Topic = topicSnap
		bundle.Network = networkcollector.Collector{}.CollectMetadata(ctx, env, networkSnap, kafkaSnap.Brokers)
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
		networkchecks.DNSChecker{},
		kafkachecks.ClusterChecker{},
		kafkachecks.RegistrationChecker{},
		kafkachecks.EndpointChecker{},
		kafkachecks.InternalTopicsChecker{},
		kraftchecks.ConfigChecker{},
		kraftchecks.ControllerChecker{},
		kraftchecks.QuorumChecker{},
		topicchecks.LeaderChecker{},
		topicchecks.ReplicaHealthChecker{},
		topicchecks.ISRChecker{MinISR: env.SelectedProfile.ExpectedMinISR},
		lintchecks.ComposeChecker{},
		lintchecks.NodeIDChecker{},
		lintchecks.ClusterIDChecker{},
		lintchecks.ProcessRolesChecker{RequireBroker: true, RequireController: true},
		lintchecks.QuorumVotersChecker{},
		lintchecks.ListenersChecker{},
		lintchecks.InterBrokerListenerChecker{},
		lintchecks.ReplicationChecker{ExpectedBrokerCount: env.SelectedProfile.BrokerCount},
		clientchecks.MetadataChecker{},
		clientchecks.ProducerChecker{},
		clientchecks.ConsumerChecker{},
		clientchecks.CommitChecker{},
		clientchecks.EndToEndChecker{},
		hostchecks.DiskChecker{},
		hostchecks.PortChecker{},
		dockerchecks.ExistenceChecker{},
		dockerchecks.RunningChecker{},
		dockerchecks.OOMChecker{},
		dockerchecks.MountChecker{},
		logchecks.SourcesChecker{},
		logchecks.FingerprintChecker{},
		logchecks.ExplanationChecker{},
		logchecks.AggregateChecker{},
	}

	results := make([]model.CheckResult, 0, len(checkers))
	for _, c := range checkers {
		results = append(results, c.Run(ctx, bundle))
	}
	return results
}
