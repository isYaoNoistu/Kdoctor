package compose

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type File struct {
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	ContainerName string      `yaml:"container_name"`
	Image         string      `yaml:"image"`
	NetworkMode   string      `yaml:"network_mode"`
	MemLimit      string      `yaml:"mem_limit"`
	Environment   Environment `yaml:"environment"`
	Volumes       []string    `yaml:"volumes"`
}

type Environment map[string]string

func (e *Environment) UnmarshalYAML(node *yaml.Node) error {
	result := map[string]string{}
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			result[node.Content[i].Value] = node.Content[i+1].Value
		}
	case yaml.SequenceNode:
		for _, item := range node.Content {
			key, value, ok := splitKV(item.Value)
			if ok {
				result[key] = value
			}
		}
	}
	*e = result
	return nil
}

func ParseFile(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read compose file: %w", err)
	}

	var file File
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse compose file: %w", err)
	}
	return &file, nil
}

func splitKV(input string) (string, string, bool) {
	for i := 0; i < len(input); i++ {
		if input[i] == '=' {
			return input[:i], input[i+1:], true
		}
	}
	return "", "", false
}
