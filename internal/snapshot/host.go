package snapshot

type HostSnapshot struct {
	Collected  bool              `json:"collected"`
	Available  bool              `json:"available"`
	DiskUsages []DiskUsage       `json:"disk_usages,omitempty"`
	PortChecks []EndpointCheck   `json:"port_checks,omitempty"`
	Errors     []string          `json:"errors,omitempty"`
	Raw        map[string]string `json:"raw,omitempty"`
}

type DiskUsage struct {
	Path           string  `json:"path"`
	TotalBytes     int64   `json:"total_bytes,omitempty"`
	AvailableBytes int64   `json:"available_bytes,omitempty"`
	UsedBytes      int64   `json:"used_bytes,omitempty"`
	UsedPercent    float64 `json:"used_percent,omitempty"`
}
