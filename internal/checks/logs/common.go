package logs

import (
	"fmt"
	"time"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

func logSnap(bundle *snapshot.Bundle) *snapshot.LogSnapshot {
	if bundle == nil {
		return nil
	}
	return bundle.Logs
}

func resultForMatchSeverity(id string, name string, summary string, severity string) model.CheckResult {
	switch severity {
	case "crit":
		return rule.NewCrit(id, name, "logs", summary)
	case "fail":
		return rule.NewFail(id, name, "logs", summary)
	case "warn":
		return rule.NewWarn(id, name, "logs", summary)
	default:
		return rule.NewPass(id, name, "logs", summary)
	}
}

func highestSeverity(matches []snapshot.LogPatternMatch) string {
	best := ""
	bestRank := -1
	for _, match := range matches {
		rank := severityRank(match.Severity)
		if rank > bestRank {
			best = match.Severity
			bestRank = rank
		}
	}
	return best
}

func severityRank(severity string) int {
	switch severity {
	case "crit":
		return 4
	case "fail":
		return 3
	case "warn":
		return 2
	default:
		return 1
	}
}

func logSourceIssues(logs *snapshot.LogSnapshot) (stale int, sparse int, empty int) {
	if logs == nil {
		return 0, 0, 0
	}
	for _, stat := range logs.SourceStats {
		if stat.Empty {
			empty++
		}
		if !stat.Fresh {
			stale++
		}
		if !stat.SufficientLines {
			sparse++
		}
	}
	return stale, sparse, empty
}

func appendSourceEvidence(result *model.CheckResult, logs *snapshot.LogSnapshot) {
	if result == nil || logs == nil {
		return
	}
	for _, stat := range logs.SourceStats {
		lastTS := "unknown"
		if stat.LastModifiedUnix > 0 {
			lastTS = time.Unix(stat.LastModifiedUnix, 0).Format(time.RFC3339)
		}
		freshness := "unknown"
		if stat.LastModifiedUnix > 0 {
			freshness = time.Since(time.Unix(stat.LastModifiedUnix, 0)).Round(time.Minute).String()
		}
		result.Evidence = append(result.Evidence, fmt.Sprintf(
			"source=%s line_count=%d byte_count=%d latest_timestamp=%s freshness=%s sample_sufficient=%t empty=%t",
			stat.Source,
			stat.Lines,
			stat.Bytes,
			lastTS,
			freshness,
			stat.SufficientLines,
			stat.Empty,
		))
	}
	for _, warning := range logs.Warnings {
		result.Evidence = append(result.Evidence, fmt.Sprintf("日志采集告警=%s", warning))
	}
	for _, err := range logs.Errors {
		result.Evidence = append(result.Evidence, fmt.Sprintf("日志采集错误=%s", err))
	}
}

func appendSourceSummary(result *model.CheckResult, logs *snapshot.LogSnapshot) {
	if result == nil || logs == nil {
		return
	}
	stale, sparse, empty := logSourceIssues(logs)
	result.Evidence = append(result.Evidence, fmt.Sprintf("source_count=%d", len(logs.SourceStats)))
	result.Evidence = append(result.Evidence, fmt.Sprintf("stale_sources=%d", stale))
	result.Evidence = append(result.Evidence, fmt.Sprintf("sparse_sources=%d", sparse))
	result.Evidence = append(result.Evidence, fmt.Sprintf("empty_sources=%d", empty))
}
