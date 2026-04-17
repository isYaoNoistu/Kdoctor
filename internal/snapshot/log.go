package snapshot

type LogSnapshot struct {
	Collected           bool              `json:"collected"`
	Available           bool              `json:"available"`
	Sources             []string          `json:"sources,omitempty"`
	SourceStats         []LogSourceStat   `json:"source_stats,omitempty"`
	Matches             []LogPatternMatch `json:"matches,omitempty"`
	Errors              []string          `json:"errors,omitempty"`
	Warnings            []string          `json:"warnings,omitempty"`
	BuiltinPatternCount int               `json:"builtin_pattern_count,omitempty"`
	CustomPatternCount  int               `json:"custom_pattern_count,omitempty"`
}

type LogSourceStat struct {
	Source           string `json:"source"`
	Kind             string `json:"kind,omitempty"`
	Lines            int    `json:"lines"`
	Bytes            int    `json:"bytes"`
	LastModifiedUnix int64  `json:"last_modified_unix,omitempty"`
	Fresh            bool   `json:"fresh"`
	SufficientLines  bool   `json:"sufficient_lines"`
	Empty            bool   `json:"empty"`
}

type LogPatternMatch struct {
	ID                string   `json:"id"`
	Library           string   `json:"library,omitempty"`
	Pattern           string   `json:"pattern"`
	Severity          string   `json:"severity"`
	Meaning           string   `json:"meaning"`
	Count             int      `json:"count"`
	Example           string   `json:"example,omitempty"`
	AffectedSources   []string `json:"affected_sources,omitempty"`
	ProbableCauses    []string `json:"probable_causes,omitempty"`
	RecommendedChecks []string `json:"recommended_checks,omitempty"`
}
