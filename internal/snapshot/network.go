package snapshot

type NetworkSnapshot struct {
	BootstrapChecks  []EndpointCheck `json:"bootstrap_checks,omitempty"`
	ControllerChecks []EndpointCheck `json:"controller_checks,omitempty"`
	MetadataChecks   []EndpointCheck `json:"metadata_checks,omitempty"`
}

type EndpointCheck struct {
	Kind       string `json:"kind"`
	Address    string `json:"address"`
	Reachable  bool   `json:"reachable"`
	DurationMs int64  `json:"duration_ms,omitempty"`
	Error      string `json:"error,omitempty"`
}
