package spec

import "gopkg.in/yaml.v3"

type Metrics struct {
	Entities []Entity `yaml:"entities"`
}

type Entity struct {
	EntityType string   `yaml:"entityType"`
	Metrics    []Metric `yaml:"metrics"`
}

type Metric struct {
	Name string `yaml:"name"`
}

func ParseMetricsFile(content []byte) (*Metrics, error) {
	specMetrics := &Metrics{}
	if err := yaml.Unmarshal(content, specMetrics); err != nil {
		return nil, err
	}
	return specMetrics, nil
}
