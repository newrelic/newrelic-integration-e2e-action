package runtime

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/newrelic"

	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
)

type MetricsTester struct {
	nrClient      newrelic.Client
	logger        *logrus.Logger
	specParentDir string
}

func NewMetricsTester(nrClient newrelic.Client, logger *logrus.Logger, specParentDir string) MetricsTester {
	return MetricsTester{
		nrClient:      nrClient,
		logger:        logger,
		specParentDir: specParentDir,
	}
}

func (mt MetricsTester) Test(tests spec.Tests, customTagKey, customTagValue string) []error {
	var errors []error
	for _, tm := range tests.Metrics {
		content, err := ioutil.ReadFile(filepath.Join(mt.specParentDir, tm.Source))
		if err != nil {
			errors = append(errors, fmt.Errorf("reading metrics source file: %w", err))
			continue
		}
		mt.logger.Debug("parsing the content of the metrics source file")
		metrics, err := spec.ParseMetricsFile(content)
		if err != nil {
			errors = append(errors, fmt.Errorf("unmarshaling metrics source file: %w", err))
			continue
		}

		queriedMetrics, err := mt.nrClient.FindEntityMetrics(dmTableName, customTagKey, customTagValue)
		if err != nil {
			errors = append(errors, fmt.Errorf("finding keyset: %w", err))
			continue
		}

		for _, entity := range metrics.Entities {
			if mt.isEntityException(entity.EntityType, tm.ExceptEntities) {
				continue
			}

			for _, metric := range entity.Metrics {
				if mt.isMetricException(metric.Name, tm.ExceptMetrics) {
					continue
				}

				if mt.containsMetric(metric.Name, queriedMetrics) {
					continue
				} else {
					errors = append(errors, fmt.Errorf("finding Metric: %v", metric.Name))
					continue
				}
			}
		}
	}
	return errors
}

func (mt MetricsTester) isEntityException(entity string, entitiesList []string) bool {
	for _, entityType := range entitiesList {
		if entityType == entity {
			return true
		}
	}
	return false
}

func (mt MetricsTester) isMetricException(metric string, exceptionMetricsList []string) bool {
	for _, exceptMetric := range exceptionMetricsList {
		if exceptMetric == metric {
			return true
		}
	}
	return false
}

func (mt MetricsTester) containsMetric(metric string, queriedMetricsList []string) bool {
	for _, queriedMetric := range queriedMetricsList {
		if queriedMetric == metric {
			return true
		}
	}
	return false
}
