package exitcode

import "kdoctor/pkg/model"

func FromReport(report model.Report) int {
	switch report.Summary.Status {
	case model.StatusCrit:
		return 3
	case model.StatusFail:
		return 2
	case model.StatusWarn:
		return 1
	case model.StatusError:
		return 5
	case model.StatusTimeout:
		return 6
	default:
		return 0
	}
}
