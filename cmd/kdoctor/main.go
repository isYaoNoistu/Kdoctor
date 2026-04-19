package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"kdoctor/internal/app"
)

func main() {
	opts, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		os.Exit(4)
	}

	application, err := app.New(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		os.Exit(4)
	}

	report, err := application.Run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		os.Exit(5)
	}

	os.Exit(report.ExitCode)
}

func parseFlags(args []string) (app.Options, error) {
	opts := app.Options{
		Mode:       "quick",
		ConfigPath: "kdoctor.yaml",
	}

	mode := "quick"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		mode = args[0]
		args = args[1:]
	}

	fs := flag.NewFlagSet("kdoctor", flag.ContinueOnError)
	fs.StringVar(&opts.ConfigPath, "config", opts.ConfigPath, "kdoctor.yaml 配置文件路径")
	fs.StringVar(&opts.ProfileName, "profile", "", "运行 profile 名称")
	fs.BoolVar(&opts.JSONOutput, "json", false, "以 JSON 格式输出")
	fs.StringVar(&opts.OutputFormat, "format", "", "输出格式：terminal/json/markdown")
	fs.StringVar(&opts.OutputPath, "output", "", "可选的输出文件路径")
	fs.StringVar(&opts.Bootstrap, "bootstrap", "", "逗号分隔的 bootstrap 地址")
	fs.StringVar(&opts.BootstrapInternal, "bootstrap-internal", "", "逗号分隔的内网 bootstrap 地址")
	fs.StringVar(&opts.BootstrapExternal, "bootstrap-external", "", "逗号分隔的外网 bootstrap 地址")
	fs.StringVar(&opts.ComposePath, "compose", "", "docker-compose 文件路径")
	fs.StringVar(&opts.LogDir, "log-dir", "", "Kafka 日志目录")
	fs.StringVar(&opts.Timeout, "timeout", "", "整体超时时间，例如 30s")
	fs.StringVar(&opts.Severity, "severity", "", "最小输出严重级别")
	fs.BoolVar(&opts.Verbose, "verbose", false, "展开 PASS/SKIP 明细与完整附录")
	if err := fs.Parse(args); err != nil {
		return app.Options{}, err
	}
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "config" {
			opts.ConfigPathExplicit = true
		}
	})

	switch mode {
	case "quick", "full", "probe", "incident", "lint":
	default:
		return app.Options{}, fmt.Errorf("不支持的模式 %q", mode)
	}
	opts.Mode = mode

	if opts.Timeout != "" {
		if _, err := time.ParseDuration(opts.Timeout); err != nil {
			return app.Options{}, fmt.Errorf("超时时间格式非法：%w", err)
		}
	}

	return opts, nil
}
