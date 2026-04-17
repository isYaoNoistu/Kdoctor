package docker

import (
	"context"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/config"
	"kdoctor/internal/snapshot"
	dockertransport "kdoctor/internal/transport/docker"
)

type Collector struct{}

func (Collector) Collect(ctx context.Context, env *config.Runtime, compose *snapshot.ComposeSnapshot) *snapshot.DockerSnapshot {
	if env == nil || !env.EnableDocker {
		return nil
	}

	expectedNames := composeutil.ContainerNames(compose, env.Config.Docker.ContainerNames)
	if len(expectedNames) == 0 {
		return nil
	}

	out := &snapshot.DockerSnapshot{
		Collected:     true,
		ExpectedNames: append([]string(nil), expectedNames...),
	}

	statusMap, err := dockertransport.ContainerStatusMap(ctx)
	if err != nil {
		out.Errors = append(out.Errors, err.Error())
		return out
	}
	out.Available = true

	for _, name := range expectedNames {
		container := snapshot.DockerContainerStatus{
			Name:   name,
			Status: statusMap[name],
		}

		inspects, err := dockertransport.InspectContainers(ctx, []string{name})
		if err != nil {
			out.Containers = append(out.Containers, container)
			out.Errors = append(out.Errors, err.Error())
			continue
		}
		if len(inspects) == 0 {
			out.Containers = append(out.Containers, container)
			continue
		}

		inspect := inspects[0]
		container.Image = inspect.Config.Image
		container.State = inspect.State.Status
		container.Status = firstNonEmpty(container.Status, inspect.State.Status)
		container.Running = inspect.State.Running
		container.RestartCount = inspect.State.RestartCount
		container.OOMKilled = inspect.State.OOMKilled
		for _, mount := range inspect.Mounts {
			container.Mounts = append(container.Mounts, snapshot.DockerMount{
				Source:      mount.Source,
				Destination: mount.Destination,
				RW:          mount.RW,
			})
		}
		out.Containers = append(out.Containers, container)
	}

	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
