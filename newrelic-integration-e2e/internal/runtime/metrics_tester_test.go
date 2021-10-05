package runtime

import (
	"io/ioutil"
	"testing"

	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsTester_Test(t *testing.T) {
	log := logrus.New()
	log.SetOutput(ioutil.Discard)
	metricsTester := NewMetricsTester(clientMock{}, log, "")

	inputTests := spec.Tests{Metrics: []spec.TestMetrics{
		{
			Source:         "testdata/powerdns.yml",
			ExceptEntities: nil,
			ExceptMetrics:  nil,
		},
	}}

	errors := metricsTester.Test(inputTests, "", "")
	assert.Equal(t, 0, len(errors))
}

func TestMetricsTester_checkMetrics(t *testing.T) {
	log := logrus.New()
	log.SetOutput(ioutil.Discard)
	metricsTester := NewMetricsTester(clientMock{}, log, "")

	entities := []spec.Entity{
		{
			EntityType: "ENTITY-A",
			Metrics: []spec.Metric{
				{"metric-A"},
			},
		},
		{
			EntityType: "ENTITY-B",
			Metrics: []spec.Metric{
				{"metric-B1"},
				{"metric-B2"},
			},
		},
	}

	tests := []struct {
		name                   string
		testMetrics            spec.TestMetrics
		queriedMetrics         []string
		numberOfErrorsExpected int
	}{

		{
			name: "when no metrics it should return 3 errors, one from each missing metric",
			testMetrics: spec.TestMetrics{
				Source:         "",
				ExceptEntities: []string{},
				ExceptMetrics:  []string{},
			},
			queriedMetrics:         []string{},
			numberOfErrorsExpected: 3,
		},
		{
			name: "when only metrics from entity B but entity A excluded it shouldn't return errors",
			testMetrics: spec.TestMetrics{
				Source:         "",
				ExceptEntities: []string{"ENTITY-A"},
				ExceptMetrics:  []string{},
			},
			queriedMetrics:         []string{"metric-B1", "metric-B2"},
			numberOfErrorsExpected: 0,
		},
		{
			name: "when a metric is not returned but it's excluded it shouldn't return errors",
			testMetrics: spec.TestMetrics{
				Source:         "",
				ExceptEntities: []string{},
				ExceptMetrics:  []string{"metric-A"},
			},
			queriedMetrics:         []string{"metric-B1", "metric-B2"},
			numberOfErrorsExpected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := metricsTester.checkMetrics(entities, tt.testMetrics, tt.queriedMetrics)
			require.Equal(t, tt.numberOfErrorsExpected, len(errors))
		})
	}
}
