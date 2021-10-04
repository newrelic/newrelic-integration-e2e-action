package runtime

import (
	"fmt"

	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/newrelic"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
	"github.com/sirupsen/logrus"
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
		err := nt.nrClient.NRQLQuery(nrql.Query, customTagKey, customTagValue)
		if err != nil {
			errors = append(errors, fmt.Errorf("querying: %w", err))
			continue
		}
	}
	return errors
}
