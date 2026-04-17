package rule

import "kdoctor/pkg/model"

func NewPass(id, name, module, summary string) model.CheckResult {
	return model.CheckResult{
		ID:       id,
		Name:     name,
		Module:   module,
		Severity: model.SeverityInfo,
		Status:   model.StatusPass,
		Summary:  summary,
	}
}

func NewWarn(id, name, module, summary string) model.CheckResult {
	return model.CheckResult{
		ID:       id,
		Name:     name,
		Module:   module,
		Severity: model.SeverityWarn,
		Status:   model.StatusWarn,
		Summary:  summary,
	}
}

func NewFail(id, name, module, summary string) model.CheckResult {
	return model.CheckResult{
		ID:       id,
		Name:     name,
		Module:   module,
		Severity: model.SeverityFail,
		Status:   model.StatusFail,
		Summary:  summary,
	}
}

func NewCrit(id, name, module, summary string) model.CheckResult {
	return model.CheckResult{
		ID:       id,
		Name:     name,
		Module:   module,
		Severity: model.SeverityCrit,
		Status:   model.StatusCrit,
		Summary:  summary,
	}
}

func NewError(id, name, module, summary, err string) model.CheckResult {
	return model.CheckResult{
		ID:           id,
		Name:         name,
		Module:       module,
		Severity:     model.SeverityFail,
		Status:       model.StatusError,
		Summary:      summary,
		ErrorMessage: err,
	}
}

func NewSkip(id, name, module, summary string) model.CheckResult {
	return model.CheckResult{
		ID:       id,
		Name:     name,
		Module:   module,
		Severity: model.SeverityInfo,
		Status:   model.StatusSkip,
		Summary:  summary,
	}
}
