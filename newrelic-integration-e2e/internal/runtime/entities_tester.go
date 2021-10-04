package runtime

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/newrelic"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
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
		guid, err := et.nrClient.FindEntityGUID(en.DataType, en.MetricName, customTagKey, customTagValue)
		if err != nil {
			errors = append(errors, fmt.Errorf("finding entity guid: %w", err))
			continue
		}
		entity, err := et.nrClient.FindEntityByGUID(guid)
		if err != nil {
			errors = append(errors, fmt.Errorf("finding entity guid: %w", err))
			continue
		}

		if entity.GetType() != en.Type {
			errors = append(errors, fmt.Errorf("entity type is not matching: %s!=%s", entity.GetType(), en.Type))
			continue
		}
	}
	return errors
}
