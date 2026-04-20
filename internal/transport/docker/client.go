package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
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

func ProcessOpenFileLimit(ctx context.Context, name string) (uint64, uint64, error) {
	output, err := shelltransport.Run(ctx, "docker", "exec", name, "sh", "-c", "cat /proc/1/limits")
	if err != nil {
		return 0, 0, fmt.Errorf("docker exec %s cat /proc/1/limits: %w", name, err)
	}
	soft, hard, parseErr := parseOpenFileLimit(output)
	if parseErr != nil {
		return 0, 0, fmt.Errorf("parse open file limit for %s: %w", name, parseErr)
	}
	return soft, hard, nil
}

func parseOpenFileLimit(input string) (uint64, uint64, error) {
	for _, line := range strings.Split(input, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 5 {
			continue
		}
		if strings.ToLower(strings.Join(fields[:3], " ")) != "max open files" {
			continue
		}
		soft, err := parseLimitValue(fields[3])
		if err != nil {
			return 0, 0, err
		}
		hard, err := parseLimitValue(fields[4])
		if err != nil {
			return 0, 0, err
		}
		return soft, hard, nil
	}
	return 0, 0, fmt.Errorf("Max open files line not found")
}

func parseLimitValue(input string) (uint64, error) {
	value := strings.TrimSpace(strings.ToLower(input))
	if value == "" {
		return 0, fmt.Errorf("empty limit value")
	}
	if value == "unlimited" {
		return math.MaxUint64, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}
