package runtime

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/newrelic/newrelic-integration-e2e-action/internal/newrelic"

	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"
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

		errors = append(errors, mt.checkMetrics(metrics.Entities, tm, queriedMetrics)...)
	}
	return errors
}

func (mt MetricsTester) checkMetrics(entities []spec.Entity, tm spec.TestMetrics, queriedMetrics []string) []error {
	var errors []error

	if tm.ExceptionsSource != "" {
		exceptMetricsPath := filepath.Join(mt.specParentDir, tm.ExceptionsSource)
		mt.logger.Debugf("parsing the content of the except metrics source file: %s", exceptMetricsPath)

		exceptions, err := parseExceptions(exceptMetricsPath)
		if err != nil {
			errors = append(errors, fmt.Errorf("reading except metrics source file %s: %w", exceptMetricsPath, err))
			return errors
		}

		tm.ExceptMetrics = append(tm.ExceptMetrics, exceptions.ExceptMetrics...)
		tm.ExceptEntities = append(tm.ExceptEntities, exceptions.ExceptEntities...)
	}

	for _, entity := range entities {
		if mt.isEntityException(entity.EntityType, tm.ExceptEntities) {
			continue
		}

		for _, metric := range entity.Metrics {
			if mt.isMetricException(metric.Name, tm.ExceptMetrics) {
				continue
			}

			if mt.containsMetric(metric.Name, queriedMetrics) {
				continue
			}

			errors = append(errors, fmt.Errorf("finding Metric: %v", metric.Name))
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

func parseExceptions(exceptMetricsPath string) (*spec.Exceptions, error) {
	content, err := ioutil.ReadFile(os.ExpandEnv(exceptMetricsPath))
	if err != nil {
		return nil, fmt.Errorf("reading except metrics source file %s: %w", exceptMetricsPath, err)
	}

	return spec.ParseExceptionsFile(content)
}
