package snapshot

type LogSnapshot struct {
	Collected bool              `json:"collected"`
	Available bool              `json:"available"`
	Sources   []string          `json:"sources,omitempty"`
	Matches   []LogPatternMatch `json:"matches,omitempty"`
	Errors    []string          `json:"errors,omitempty"`
}

type LogPatternMatch struct {
	ID                string   `json:"id"`
	Pattern           string   `json:"pattern"`
	Severity          string   `json:"severity"`
	Meaning           string   `json:"meaning"`
	Count             int      `json:"count"`
	Example           string   `json:"example,omitempty"`
	AffectedSources   []string `json:"affected_sources,omitempty"`
	ProbableCauses    []string `json:"probable_causes,omitempty"`
	RecommendedChecks []string `json:"recommended_checks,omitempty"`
}
