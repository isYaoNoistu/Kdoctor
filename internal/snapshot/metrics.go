package snapshot

type MetricsSnapshot struct {
	Collected bool                    `json:"collected"`
	Available bool                    `json:"available"`
	Path      string                  `json:"path,omitempty"`
	Endpoints []MetricsEndpointStatus `json:"endpoints,omitempty"`
	Errors    []string                `json:"errors,omitempty"`
}

type MetricsEndpointStatus struct {
	Name           string             `json:"name,omitempty"`
	Address        string             `json:"address"`
	Reachable      bool               `json:"reachable"`
	DurationMs     int64              `json:"duration_ms,omitempty"`
	ServerTimeUnix int64              `json:"server_time_unix,omitempty"`
	Error          string             `json:"error,omitempty"`
	Metrics        map[string]float64 `json:"metrics,omitempty"`
}
