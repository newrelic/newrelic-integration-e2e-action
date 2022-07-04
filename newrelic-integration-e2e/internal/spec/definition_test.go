package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseExceptionsFile(t *testing.T) {
	sample := `
except_metrics:
- metric_a
- metric_b
except_entities:
- entity_a
`
	exceptions, err := ParseExceptionsFile([]byte(sample))
	assert.Nil(t, err)
	assert.Equal(
		t,
		&Exceptions{
			ExceptEntities: []string{"entity_a"},
			ExceptMetrics:  []string{"metric_a", "metric_b"},
		},
		exceptions)
}

func Test_ParseDefinitionFile(t *testing.T) {
	sample := `
description: |
  End-to-end tests for PowerDNS integration

agent:
  build_context: /path/to/compose/dir
  integrations:
    nri-prometheus:  bin/nri-prometheus
  env_vars:
    NRJMX_VERSION: "1.5.3"
scenarios:
  - description: |
      Scenario Description.
    before:
      - docker-compose -f deps/docker-compose.yml up -d
    after:
      - docker-compose -f deps/docker-compose.yml down -v
    integrations:
      - name: nri-powerdns
        binary_path: bin/nri-powerdns
        exporter_binary_path: bin/nri-powerdns-exporter
        config:
          powerdns_url: http://localhost:8081/api/v1/
    tests:
      nrqls:
        - query: "a-query"
      entities:
        - type: "POWERDNS_AUTHORITATIVE"
          data_type: "Metric"
          metric_name: "powerdns_authoritative_up"
      metrics:
        - source: "powerdns.yml"
          except_metrics:
            - powerdns_authoritative_answers_bytes_total`
	spec, err := ParseDefinitionFile([]byte(sample))
	assert.Nil(t, err)
	assert.Equal(t, "End-to-end tests for PowerDNS integration\n", spec.Description)

	expectedAgentExtensions := Agent{
		BuildContext: "/path/to/compose/dir",
		Integrations: map[string]string{
			"nri-prometheus": "bin/nri-prometheus",
		},
		EnvVars: map[string]string{
			"NRJMX_VERSION": "1.5.3",
		},
	}
	assert.Equal(t, &expectedAgentExtensions, spec.AgentExtensions)

	expectedScenarios := []Scenario{
		{
			Description: "Scenario Description.\n",
			Integrations: []Integration{
				{
					Name:               "nri-powerdns",
					BinaryPath:         "bin/nri-powerdns",
					ExporterBinaryPath: "bin/nri-powerdns-exporter",
					Config: map[string]interface{}{
						"powerdns_url": "http://localhost:8081/api/v1/",
					},
				},
			},
			Before: []string{"docker-compose -f deps/docker-compose.yml up -d"},
			After:  []string{"docker-compose -f deps/docker-compose.yml down -v"},
			Tests: Tests{
				NRQLs: []TestNRQL{{Query: "a-query"}},
				Entities: []TestEntity{
					{
						Type:       "POWERDNS_AUTHORITATIVE",
						DataType:   "Metric",
						MetricName: "powerdns_authoritative_up",
					},
				},
				Metrics: []TestMetrics{
					{
						Source: "powerdns.yml",
						Exceptions: Exceptions{
							ExceptMetrics: []string{"powerdns_authoritative_answers_bytes_total"},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, expectedScenarios, spec.Scenarios)
}
