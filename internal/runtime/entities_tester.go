package runtime

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/newrelic/newrelic-integration-e2e-action/internal/newrelic"
	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"
)

type EntitiesTester struct {
	nrClient newrelic.Client
	logger   *logrus.Logger
}

func NewEntitiesTester(nrClient newrelic.Client, logger *logrus.Logger) EntitiesTester {
	return EntitiesTester{
		nrClient: nrClient,
		logger:   logger,
	}
}

func (et EntitiesTester) Test(tests spec.Tests, customTagKey, customTagValue string) []error {
	var errors []error
	for _, en := range tests.Entities {
		// By default if not notified, we set expectedNumber to 1
		if en.ExpectedNumber == 0 {
			en.ExpectedNumber = 1
		}
		guids, err := et.nrClient.FindEntityGUIDs(en.DataType, en.MetricName, customTagKey, customTagValue, en.ExpectedNumber)
		if err != nil {
			errors = append(errors, fmt.Errorf("finding entity guid: %w", err))
			continue
		}
		for _, guid := range guids {
			entity, err := et.nrClient.FindEntityByGUID(&guid)
			if err != nil {
				errors = append(errors, fmt.Errorf("finding entity guid: %w", err))
				continue
			}

			// Some entity GUIDs (from sample shimming) don't return any object, if it's the case we don't fail the test
			if entity != nil && entity.GetType() != en.Type {
				errors = append(errors, fmt.Errorf("entity type is not matching: %s!=%s", entity.GetType(), en.Type))
				continue
			}
		}
	}
	return errors
}
