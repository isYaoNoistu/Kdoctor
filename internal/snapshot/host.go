package snapshot

type HostSnapshot struct {
	Collected           bool              `json:"collected"`
	Available           bool              `json:"available"`
	DiskUsages          []DiskUsage       `json:"disk_usages,omitempty"`
	PortChecks          []EndpointCheck   `json:"port_checks,omitempty"`
	ObservedListenPorts []int             `json:"observed_listen_ports,omitempty"`
	FD                  *FDStats          `json:"fd,omitempty"`
	ContainerFD         []ContainerFDStat `json:"container_fd,omitempty"`
	Memory              *MemoryStats      `json:"memory,omitempty"`
	Errors              []string          `json:"errors,omitempty"`
	Raw                 map[string]string `json:"raw,omitempty"`
}

type DiskUsage struct {
	Path            string  `json:"path"`
	TotalBytes      int64   `json:"total_bytes,omitempty"`
	AvailableBytes  int64   `json:"available_bytes,omitempty"`
	UsedBytes       int64   `json:"used_bytes,omitempty"`
	UsedPercent     float64 `json:"used_percent,omitempty"`
	TotalInodes     int64   `json:"total_inodes,omitempty"`
	AvailableInodes int64   `json:"available_inodes,omitempty"`
	UsedInodes      int64   `json:"used_inodes,omitempty"`
	UsedInodePct    float64 `json:"used_inode_pct,omitempty"`
}

type FDStats struct {
	SoftLimit  uint64 `json:"soft_limit,omitempty"`
	SystemUsed uint64 `json:"system_used,omitempty"`
	SystemMax  uint64 `json:"system_max,omitempty"`
}

type ContainerFDStat struct {
	Name      string `json:"name"`
	SoftLimit uint64 `json:"soft_limit,omitempty"`
	HardLimit uint64 `json:"hard_limit,omitempty"`
	Error     string `json:"error,omitempty"`
}

type MemoryStats struct {
	TotalBytes     int64   `json:"total_bytes,omitempty"`
	AvailableBytes int64   `json:"available_bytes,omitempty"`
	UsedBytes      int64   `json:"used_bytes,omitempty"`
	UsedPercent    float64 `json:"used_percent,omitempty"`
}
