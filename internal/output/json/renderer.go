package json

import (
	"encoding/json"

	"kdoctor/pkg/model"
)

type Renderer struct{}

func (Renderer) Render(report model.Report) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}
