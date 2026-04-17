package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	shelltransport "kdoctor/internal/transport/shell"
)

type Inspect struct {
	Name   string `json:"Name"`
	Config struct {
		Image string `json:"Image"`
	} `json:"Config"`
	State struct {
		Status       string `json:"Status"`
		Running      bool   `json:"Running"`
		RestartCount int    `json:"RestartCount"`
		OOMKilled    bool   `json:"OOMKilled"`
	} `json:"State"`
	Mounts []struct {
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		RW          bool   `json:"RW"`
	} `json:"Mounts"`
}

func InspectContainers(ctx context.Context, names []string) ([]Inspect, error) {
	if len(names) == 0 {
		return nil, nil
	}
	args := append([]string{"inspect"}, names...)
	output, err := shelltransport.Run(ctx, "docker", args...)
	if err != nil {
		return nil, fmt.Errorf("docker inspect: %w", err)
	}
	var out []Inspect
	if err := json.Unmarshal([]byte(output), &out); err != nil {
		return nil, fmt.Errorf("decode docker inspect output: %w", err)
	}
	return out, nil
}

func ContainerStatusMap(ctx context.Context) (map[string]string, error) {
	output, err := shelltransport.Run(ctx, "docker", "ps", "-a", "--format", "{{.Names}}|{{.Status}}")
	if err != nil {
		return nil, fmt.Errorf("docker ps: %w", err)
	}
	statuses := map[string]string{}
	if strings.TrimSpace(output) == "" {
		return statuses, nil
	}
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "|", 2)
		if len(parts) != 2 {
			continue
		}
		statuses[parts[0]] = parts[1]
	}
	return statuses, nil
}

func Logs(ctx context.Context, name string, tail int, since string) (string, error) {
	args := []string{"logs"}
	if strings.TrimSpace(since) != "" {
		args = append(args, "--since", since)
	}
	args = append(args, "--tail", fmt.Sprintf("%d", tail), name)
	return shelltransport.Run(ctx, "docker", args...)
}
