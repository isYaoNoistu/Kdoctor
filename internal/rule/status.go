package rule

import "kdoctor/pkg/model"

func StatusRank(status model.CheckStatus) int {
	switch status {
	case model.StatusCrit:
		return 7
	case model.StatusFail:
		return 6
	case model.StatusWarn:
		return 5
	case model.StatusError:
		return 4
	case model.StatusTimeout:
		return 3
	case model.StatusSkip:
		return 2
	case model.StatusPass:
		return 1
	default:
		return 0
	}
}
