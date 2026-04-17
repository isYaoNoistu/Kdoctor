package model

type CheckResult struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Module         string         `json:"module"`
	Severity       CheckSeverity  `json:"severity"`
	Status         CheckStatus    `json:"status"`
	Target         string         `json:"target,omitempty"`
	Summary        string         `json:"summary"`
	Evidence       []string       `json:"evidence,omitempty"`
	Impact         string         `json:"impact,omitempty"`
	PossibleCauses []string       `json:"possible_causes,omitempty"`
	NextActions    []string       `json:"next_actions,omitempty"`
	DurationMs     int64          `json:"duration_ms,omitempty"`
	ErrorMessage   string         `json:"error_message,omitempty"`
	Raw            map[string]any `json:"raw,omitempty"`
}

func (c CheckResult) IsProblem() bool {
	switch c.Status {
	case StatusWarn, StatusFail, StatusCrit, StatusError, StatusTimeout:
		return true
	default:
		return false
	}
}
