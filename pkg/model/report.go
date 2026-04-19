package model

import (
	"sort"
	"time"

	"kdoctor/pkg/buildinfo"
)

type Summary struct {
	Status             CheckStatus `json:"status"`
	BrokerTotal        int         `json:"broker_total"`
	BrokerAlive        int         `json:"broker_alive"`
	ControllerOK       bool        `json:"controller_ok"`
	CriticalCount      int         `json:"critical_count"`
	FailCount          int         `json:"fail_count"`
	WarnCount          int         `json:"warn_count"`
	ErrorCount         int         `json:"error_count"`
	SkipCount          int         `json:"skip_count"`
	Overview           string      `json:"overview,omitempty"`
	DataSourceCoverage []string    `json:"data_source_coverage,omitempty"`
	DegradedTasks      []string    `json:"degraded_tasks,omitempty"`
	RootCauses         []string    `json:"root_causes,omitempty"`
	RecommendedActions []string    `json:"recommended_actions,omitempty"`
}

type Report struct {
	SchemaVersion string        `json:"schema_version"`
	ToolVersion   string        `json:"tool_version"`
	Mode          string        `json:"mode"`
	Profile       string        `json:"profile"`
	ExitCode      int           `json:"exit_code"`
	CheckedAt     time.Time     `json:"checked_at"`
	ElapsedMs     int64         `json:"elapsed_ms"`
	Summary       Summary       `json:"summary"`
	Checks        []CheckResult `json:"checks"`
	Errors        []string      `json:"errors,omitempty"`
}

func NewReport(mode, profile string, checkedAt time.Time) Report {
	return Report{
		SchemaVersion: "kdoctor.report.v2",
		ToolVersion:   buildinfo.ToolVersion(),
		Mode:          mode,
		Profile:       profile,
		CheckedAt:     checkedAt,
		Summary: Summary{
			Status: StatusPass,
		},
	}
}

func (r *Report) AddCheck(check CheckResult) {
	r.Checks = append(r.Checks, check)
}

func (r *Report) AddError(err string) {
	if err == "" {
		return
	}
	r.Errors = append(r.Errors, err)
}

func (r *Report) Finalize() {
	sort.SliceStable(r.Checks, func(i, j int) bool {
		left := statusRank(r.Checks[i].Status)
		right := statusRank(r.Checks[j].Status)
		if left != right {
			return left > right
		}
		if r.Checks[i].Module != r.Checks[j].Module {
			return r.Checks[i].Module < r.Checks[j].Module
		}
		return r.Checks[i].ID < r.Checks[j].ID
	})

	var summary Summary
	summary.Status = StatusPass
	for i := range r.Checks {
		check := r.Checks[i]
		check.Evidence = normalizeStrings(check.Evidence)
		check.PossibleCauses = normalizeStrings(check.PossibleCauses)
		check.NextActions = normalizeStrings(check.NextActions)
		r.Checks[i] = check
		switch check.Status {
		case StatusCrit:
			summary.CriticalCount++
		case StatusFail:
			summary.FailCount++
		case StatusWarn:
			summary.WarnCount++
		case StatusError, StatusTimeout:
			summary.ErrorCount++
		case StatusSkip:
			summary.SkipCount++
		}

		if statusRank(check.Status) > statusRank(summary.Status) {
			summary.Status = check.Status
		}
	}
	if len(r.Errors) > 0 && summary.Status == StatusPass {
		summary.Status = StatusError
	}
	summary.BrokerTotal = r.Summary.BrokerTotal
	summary.BrokerAlive = r.Summary.BrokerAlive
	summary.ControllerOK = r.Summary.ControllerOK
	summary.Overview = r.Summary.Overview
	summary.DataSourceCoverage = r.Summary.DataSourceCoverage
	summary.DegradedTasks = r.Summary.DegradedTasks
	summary.RootCauses = r.Summary.RootCauses
	summary.RecommendedActions = r.Summary.RecommendedActions
	r.Summary = summary
}

func normalizeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func statusRank(status CheckStatus) int {
	switch status {
	case StatusCrit:
		return 7
	case StatusFail:
		return 6
	case StatusWarn:
		return 5
	case StatusError:
		return 4
	case StatusTimeout:
		return 3
	case StatusSkip:
		return 2
	case StatusPass:
		return 1
	default:
		return 0
	}
}
