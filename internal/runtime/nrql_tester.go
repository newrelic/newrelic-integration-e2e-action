package runtime

import (
	"errors"
	"fmt"
	"github.com/newrelic/newrelic-integration-e2e-action/internal/newrelic"
	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"
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

var ErrorExpected = errors.New("an error was expected")

func (nt NRQLTester) Test(tests spec.Tests, customTagKey, customTagValue string) []error {
	var errors []error
	for _, nrql := range tests.NRQLs {
		err := nt.nrClient.NRQLQuery(nrql.Query, customTagKey, customTagValue, nrql.ErrorExpected, nrql.ExpectedResults)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s", err.Error()))
		}
	}
	return errors
}
