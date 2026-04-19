package json

import (
	"encoding/json"
	"testing"
	"time"

	"kdoctor/pkg/model"
)

func TestRendererIncludesSchemaAndToolVersion(t *testing.T) {
	report := model.NewReport(model.ModeProbe, "generic-bootstrap", time.Date(2026, 4, 19, 21, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	report.Finalize()

	payload, err := Renderer{}.Render(report)
	if err != nil {
		t.Fatalf("render json: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode json: %v", err)
	}

	if decoded["schema_version"] != "kdoctor.report.v2" {
		t.Fatalf("unexpected schema_version: %#v", decoded["schema_version"])
	}
	if decoded["tool_version"] == "" {
		t.Fatal("expected tool_version to be present")
	}
}
