package compose

import (
	"context"

	"kdoctor/internal/config"
	parsecompose "kdoctor/internal/parser/compose"
	"kdoctor/internal/snapshot"
)

type Collector struct{}

func (Collector) Collect(_ context.Context, env *config.Runtime) (*snapshot.ComposeSnapshot, error) {
	if env.ComposePath == "" {
		return nil, nil
	}

	file, err := parsecompose.ParseFile(env.ComposePath)
	if err != nil {
		return nil, err
	}

	out := &snapshot.ComposeSnapshot{
		SourcePath: env.ComposePath,
		Services:   map[string]snapshot.ComposeService{},
	}
	for name, service := range file.Services {
		out.Services[name] = snapshot.ComposeService{
			Name:          name,
			ContainerName: service.ContainerName,
			Image:         service.Image,
			NetworkMode:   service.NetworkMode,
			MemLimit:      service.MemLimit,
			Environment:   map[string]string(service.Environment),
			Volumes:       append([]string(nil), service.Volumes...),
		}
	}
	return out, nil
}
