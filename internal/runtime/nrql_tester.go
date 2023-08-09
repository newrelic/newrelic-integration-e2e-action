package runtime

import (
	"errors"
	"fmt"
	"github.com/newrelic/newrelic-integration-e2e-action/internal/newrelic"
	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"
	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidConfig = errors.New("invalid NRQL test config")
)

type NRQLTester struct {
	nrClient newrelic.Client
	logger   *logrus.Logger
}

func NewNRQLTester(nrClient newrelic.Client, logger *logrus.Logger) NRQLTester {
	return NRQLTester{
		nrClient: nrClient,
		logger:   logger,
	}
}

func (nt NRQLTester) Test(tests spec.Tests, customTagKey, customTagValue string) []error {
	var errors []error
	for _, nrql := range tests.NRQLs {
		configErr := validateNRQLTestConfig(nrql)
		if configErr == nil {
			testErr := nt.nrClient.NRQLQuery(nrql.Query, customTagKey, customTagValue, nrql.ErrorExpected, nrql.ExpectedResults)
			if testErr != nil {
				errors = append(errors, fmt.Errorf("%w", testErr))
			}
		} else {
			errors = append(errors, configErr)
		}

	}
	return errors
}

func validateNRQLTestConfig(nrqlTest spec.TestNRQL) error {
	if nrqlTest.Query == "" {
		return fmt.Errorf("%w: missing query param", ErrInvalidConfig)
	}

	if nrqlTest.ExpectedResults != nil {
		// Check expected value config
		if nrqlTest.ErrorExpected {
			return fmt.Errorf("%w: expected_results cannot be used with error_expected", ErrInvalidConfig)
		}

		for i, expectedResult := range nrqlTest.ExpectedResults {
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
		}
	}

	return nil
}
