package runner

import (
	"context"
	"time"

	"kdoctor/internal/config"
	"kdoctor/internal/diagnose"
	"kdoctor/pkg/model"
)

type Runner struct{}

func New() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, env *config.Runtime) (model.Report, error) {
	startedAt := time.Now()
	bundle, checks, errs := CollectAndCheck(ctx, env)

	report := model.NewReport(env.Mode, env.ProfileName, startedAt)
	if bundle != nil && bundle.Kafka != nil {
		report.Summary.BrokerTotal = bundle.Kafka.ExpectedBrokerCount
		if report.Summary.BrokerTotal == 0 {
			report.Summary.BrokerTotal = len(bundle.Kafka.Brokers)
		}
		report.Summary.BrokerAlive = len(bundle.Kafka.Brokers)
		report.Summary.ControllerOK = bundle.Kafka.ControllerID != nil
	}
	for _, check := range checks {
		report.AddCheck(check)
	}
	for _, err := range errs {
		report.AddError(err)
	}
	report.ElapsedMs = time.Since(startedAt).Milliseconds()
	report.Finalize()
	diagnose.RootCause{
		MaxCauses:        env.DiagnosisMaxRootCauses,
		EnableConfidence: env.DiagnosisEnableConfidence,
	}.Diagnose(&report)
	diagnose.Incident{}.Summarize(&report)
	return report, nil
}
