package snapshot

type ComposeSnapshot struct {
	SourcePath string                    `json:"source_path,omitempty"`
	Services   map[string]ComposeService `json:"services,omitempty"`
}

type ComposeService struct {
	Name          string            `json:"name"`
	ContainerName string            `json:"container_name,omitempty"`
	Image         string            `json:"image,omitempty"`
	NetworkMode   string            `json:"network_mode,omitempty"`
	MemLimit      string            `json:"mem_limit,omitempty"`
	Environment   map[string]string `json:"environment,omitempty"`
	Volumes       []string          `json:"volumes,omitempty"`
}
