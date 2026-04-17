package docker

import "kdoctor/internal/snapshot"

func dockerSnap(bundle *snapshot.Bundle) *snapshot.DockerSnapshot {
	if bundle == nil {
		return nil
	}
	return bundle.Docker
}

func dockerContainerMap(docker *snapshot.DockerSnapshot) map[string]snapshot.DockerContainerStatus {
	out := map[string]snapshot.DockerContainerStatus{}
	if docker == nil {
		return out
	}
	for _, container := range docker.Containers {
		out[container.Name] = container
	}
	return out
}
