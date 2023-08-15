package spec

import (
	"errors"
	"fmt"
	yaml "gopkg.in/yaml.v3"
)

var (
	ErrInvalidConfig = errors.New("invalid NRQL test config")
)

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
	Query           string                   `yaml:"query"`
	ErrorExpected   bool                     `yaml:"error_expected"`
	ExpectedResults []TestNRQLExpectedResult `yaml:"expected_results"`
}

type TestNRQLExpectedResult struct {
	Key               string   `yaml:"key"`
	Value             any      `yaml:"value"`
	LowerBoundedValue *float64 `yaml:"lowerBoundedValue"`
	UpperBoundedValue *float64 `yaml:"upperBoundedValue"`
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

	for _, scenario := range specDefinition.Scenarios {
		for _, nrql := range scenario.Tests.NRQLs {
			err := validateNRQLTestConfig(nrql)
			if err != nil {
				return nil, err
			}
		}
	}

	if specDefinition.CustomTestKey == "" {
		specDefinition.CustomTestKey = defaultCustomTagKey
	}

	return specDefinition, nil
}

func validateNRQLTestConfig(nrqlTest TestNRQL) error {
	if nrqlTest.Query == "" {
		return fmt.Errorf("%w: missing query param", ErrInvalidConfig)
	}

	if nrqlTest.ExpectedResults != nil {
		// Check expected value config
		if nrqlTest.ErrorExpected {
			return fmt.Errorf("%w: expected_results cannot be used with error_expected", ErrInvalidConfig)
		}

		for i, expectedResult := range nrqlTest.ExpectedResults {
			err := validateExpectedResult(expectedResult, i)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func validateExpectedResult(expectedResult TestNRQLExpectedResult, i int) error {
	if expectedResult.Value != nil {
		// Ensure bounds are nil
		if expectedResult.LowerBoundedValue != nil || expectedResult.UpperBoundedValue != nil {
			return fmt.Errorf("%w: expected_results[%d].value cannot be used with bounded expected values", ErrInvalidConfig, i)
		}
	} else {
		if expectedResult.LowerBoundedValue == nil && expectedResult.UpperBoundedValue == nil {
			return fmt.Errorf("%w: at least 1 expected value bound is required when not using expected_results[%d].value", ErrInvalidConfig, i)
		}
	}
	return nil
}
