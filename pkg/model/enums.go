package model

type CheckStatus string

const (
	StatusPass    CheckStatus = "PASS"
	StatusWarn    CheckStatus = "WARN"
	StatusFail    CheckStatus = "FAIL"
	StatusCrit    CheckStatus = "CRIT"
	StatusError   CheckStatus = "ERROR"
	StatusSkip    CheckStatus = "SKIP"
	StatusTimeout CheckStatus = "TIMEOUT"
)

type CheckSeverity string

const (
	SeverityInfo CheckSeverity = "INFO"
	SeverityWarn CheckSeverity = "WARN"
	SeverityFail CheckSeverity = "FAIL"
	SeverityCrit CheckSeverity = "CRIT"
)

const (
	ModeQuick    = "quick"
	ModeFull     = "full"
	ModeProbe    = "probe"
	ModeIncident = "incident"
	ModeLint     = "lint"
)
