package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"kdoctor/internal/config"
	"kdoctor/internal/exitcode"
	"kdoctor/internal/localize"
	jsonoutput "kdoctor/internal/output/json"
	markdownoutput "kdoctor/internal/output/markdown"
	"kdoctor/internal/output/terminal"
	"kdoctor/internal/profile"
	"kdoctor/internal/runner"
	"kdoctor/pkg/model"
)

type Options struct {
	Mode              string
	ConfigPath        string
	ProfileName       string
	JSONOutput        bool
	OutputFormat      string
	OutputPath        string
	Bootstrap         string
	BootstrapInternal string
	BootstrapExternal string
	ComposePath       string
	LogDir            string
	Timeout           string
	Severity          string
}

type App struct {
	env     *config.Runtime
	runner  *runner.Runner
	options Options
}

func New(opts Options) (*App, error) {
	env, err := Bootstrap(opts)
	if err != nil {
		return nil, err
	}
	return &App{
		env:     env,
		runner:  runner.New(),
		options: opts,
	}, nil
}

func (a *App) Run(ctx context.Context) (model.Report, error) {
	ctx, cancel := context.WithTimeout(ctx, a.env.Timeout)
	defer cancel()

	report, err := a.runner.Run(ctx, a.env)
	if err != nil {
		return model.Report{}, err
	}

	report.ExitCode = exitcode.FromReport(report)
	localize.ApplyChinese(&report)
	payload, err := a.render(report)
	if err != nil {
		return model.Report{}, fmt.Errorf("render report: %w", err)
	}

	if a.options.OutputPath != "" {
		if err := os.WriteFile(a.options.OutputPath, payload, 0o644); err != nil {
			return model.Report{}, fmt.Errorf("write report file: %w", err)
		}
	}

	fmt.Print(string(payload))
	return report, nil
}

func (a *App) render(report model.Report) ([]byte, error) {
	switch localize.GuessFormat(a.options.OutputFormat, a.options.OutputPath, a.options.JSONOutput) {
	case "json":
		return jsonoutput.Renderer{}.Render(report)
	case "markdown", "md":
		return markdownoutput.Renderer{}.Render(report)
	default:
		return terminal.Renderer{}.Render(report)
	}
}

func parseCSV(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseDurationOrDefault(input string, fallback string) (time.Duration, error) {
	if strings.TrimSpace(input) == "" {
		input = fallback
	}
	return time.ParseDuration(input)
}

func availableProfileName(opts Options, fileCfg config.Config) string {
	if opts.ProfileName != "" {
		return opts.ProfileName
	}
	if fileCfg.DefaultProfile != "" {
		return fileCfg.DefaultProfile
	}
	return config.Default().DefaultProfile
}

func mergeProfileConfig(base config.Config, selected string) config.Config {
	cfg := profile.ApplyBuiltin(base, selected)
	if profileCfg, ok := cfg.Profiles[selected]; ok {
		if cfg.Profiles == nil {
			cfg.Profiles = map[string]config.ProfileConfig{}
		}
		cfg.Profiles[selected] = profileCfg
	}
	return cfg
}
