package snapshot

type DockerSnapshot struct {
	ExpectedNames []string                `json:"expected_names,omitempty"`
	Collected     bool                    `json:"collected"`
	Available     bool                    `json:"available"`
	Containers    []DockerContainerStatus `json:"containers,omitempty"`
	Errors        []string                `json:"errors,omitempty"`
}

type DockerContainerStatus struct {
	Name         string        `json:"name"`
	Image        string        `json:"image,omitempty"`
	State        string        `json:"state,omitempty"`
	Status       string        `json:"status,omitempty"`
	Running      bool          `json:"running"`
	RestartCount int           `json:"restart_count,omitempty"`
	OOMKilled    bool          `json:"oom_killed"`
	Mounts       []DockerMount `json:"mounts,omitempty"`
}

type DockerMount struct {
	Source      string `json:"source,omitempty"`
	Destination string `json:"destination,omitempty"`
	RW          bool   `json:"rw"`
}
