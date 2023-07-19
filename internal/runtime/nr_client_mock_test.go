package runtime

import (
	"errors"
	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"

	"github.com/newrelic/newrelic-client-go/pkg/common"
	"github.com/newrelic/newrelic-client-go/pkg/entities"
)

const (
	errFindEntityGUID   = "wrongEntitySample"
	errFindEntityByGUID = "wrongEntityGUID"
	correctEntityType   = "correctEntityType"
	errNRQLQuery        = "wrongNRQLQuery"
)

var (
	ErrorTest = errors.New("an-error")
)

type clientMock struct{}

func (c clientMock) FindEntityGUIDs(sample, metricName, customTagKey, entityTag string, expectedNumber int) ([]common.EntityGUID, error) {
	switch sample {
	case errFindEntityGUID:
		return nil, ErrorTest
	case errFindEntityByGUID:
		guid := common.EntityGUID(errFindEntityByGUID)
		return []common.EntityGUID{guid}, nil
	}

	guid := common.EntityGUID("AAAA")
	return []common.EntityGUID{guid}, nil
}

func (c clientMock) FindEntityByGUID(guid *common.EntityGUID) (entities.EntityInterface, error) {
	if *guid == errFindEntityByGUID {
		return nil, ErrorTest
	}
	return entities.EntityInterface(&entities.GenericInfrastructureEntity{Type: correctEntityType}), nil
}

func (c clientMock) FindEntityMetrics(sample, customTagKey, entityTag string) ([]string, error) {
	return []string{"powerdns_authoritative_deferred_cache_actions"}, nil
}

func (c clientMock) NRQLQuery(query, customTagKey, entityTag string, errorExpected bool, expectedResults []spec.TestNRQLExpectedResult) error {
	if query == errNRQLQuery && !errorExpected {
		return ErrorTest
	}
	if query != errNRQLQuery && errorExpected {
		return ErrorTest
	}
	return nil
}
