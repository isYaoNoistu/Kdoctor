package logs

import (
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
