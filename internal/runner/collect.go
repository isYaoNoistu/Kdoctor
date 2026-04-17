package runner

import (
	"context"

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
	var composeErr error
	var networkSnap *snapshot.NetworkSnapshot

	parallel(
		func() {
			composeSnap, composeErr = composecollector.Collector{}.Collect(ctx, env)
		},
		func() {
			networkSnap = networkcollector.Collector{}.CollectBase(ctx, env)
		},
	)

	if composeErr != nil {
		errs = append(errs, composeErr.Error())
	}
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

	checks := runChecks(ctx, env, bundle)
	return bundle, checks, errs
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
