package exitcode

import (
	"testing"
	"time"

	"kdoctor/pkg/model"
)

func TestFromReport(t *testing.T) {
	report := model.NewReport("quick", "test", time.Now())
	report.Summary.Status = model.StatusCrit
	if got := FromReport(report); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}

	report.Summary.Status = model.StatusFail
	if got := FromReport(report); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}

	report.Summary.Status = model.StatusWarn
	if got := FromReport(report); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}
