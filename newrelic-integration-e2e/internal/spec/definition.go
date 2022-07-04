package spec

import yaml "gopkg.in/yaml.v3"

const defaultCustomTagKey = "testKey"

type Definition struct {
	Description     string     `yaml:"description"`
	Scenarios       []Scenario `yaml:"scenarios"`
	AgentExtensions *Agent     `yaml:"agent"`
	PlainLogs       bool       `yaml:"plain_logs"`
	CustomTestKey   string     `yaml:"custom_test_key"`
}

type Agent struct {
	BuildContext string            `yaml:"build_context"`
	Integrations map[string]string `yaml:"integrations"`
	EnvVars      map[string]string `yaml:"env_vars"`
}

type Scenario struct {
	Description  string        `yaml:"description"`
	Integrations []Integration `yaml:"integrations"`
	Before       []string      `yaml:"before"`
	After        []string      `yaml:"after"`
	Tests        Tests         `yaml:"tests"`
}

type Integration struct {
	Name               string                 `yaml:"name"`
	BinaryPath         string                 `yaml:"binary_path"`
	ExporterBinaryPath string                 `yaml:"exporter_binary_path"`
	Config             map[string]interface{} `yaml:"config"`
	Env                map[string]interface{} `yaml:"env"`
}

type Tests struct {
	NRQLs    []TestNRQL    `yaml:"nrqls"`
	Entities []TestEntity  `yaml:"entities"`
	Metrics  []TestMetrics `yaml:"metrics"`
}

type TestNRQL struct {
	Query string `yaml:"query"`
}

type TestEntity struct {
	Type           string `yaml:"type"`
	DataType       string `yaml:"data_type"`
	MetricName     string `yaml:"metric_name"`
	ExpectedNumber int    `yaml:"expected_number"`
}

type TestMetrics struct {
	Source           string `yaml:"source"`
	ExceptionsSource string `yaml:"exceptions_source"`
	Exceptions       `yaml:",inline"`
}

type Exceptions struct {
	ExceptEntities []string `yaml:"except_entities"`
	ExceptMetrics  []string `yaml:"except_metrics"`
}

func ParseExceptionsFile(content []byte) (*Exceptions, error) {
	exceptions := &Exceptions{}

	if err := yaml.Unmarshal(content, exceptions); err != nil {
		return nil, err
	}

	return exceptions, nil
}

func ParseDefinitionFile(content []byte) (*Definition, error) {
	specDefinition := &Definition{}
	if err := yaml.Unmarshal(content, specDefinition); err != nil {
		return nil, err
	}

	if specDefinition.CustomTestKey == "" {
		specDefinition.CustomTestKey = defaultCustomTagKey
	}

	return specDefinition, nil
}
