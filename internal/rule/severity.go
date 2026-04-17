package rule

import "kdoctor/pkg/model"

func SeverityRank(severity model.CheckSeverity) int {
	switch severity {
	case model.SeverityCrit:
		return 4
	case model.SeverityFail:
		return 3
	case model.SeverityWarn:
		return 2
	case model.SeverityInfo:
		return 1
	default:
		return 0
	}
}
