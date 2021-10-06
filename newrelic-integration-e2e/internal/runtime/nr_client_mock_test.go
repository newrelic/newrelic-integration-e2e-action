package runtime

import (
	"errors"

	"github.com/newrelic/newrelic-client-go/pkg/entities"
)

const (
	errFindEntityGUID   = "wrongEntitySample"
	errFindEntityByGUID = "wrongEntityGUID"
	correctEntityType   = "correctEntityType"
	errNRQLQuery        = "wrongNRQLQuery"
)

type clientMock struct{}

func (c clientMock) FindEntityGUID(sample, metricName, customTagKey, entityTag string) (*entities.EntityGUID, error) {
	switch sample {
	case errFindEntityGUID:
		return nil, errors.New("an-error")
	case errFindEntityByGUID:
		guid := entities.EntityGUID(errFindEntityByGUID)
		return &guid, nil
	}

	guid := entities.EntityGUID("AAAA")
	return &guid, nil
}

func (c clientMock) FindEntityByGUID(guid *entities.EntityGUID) (entities.EntityInterface, error) {
	if *guid == errFindEntityByGUID {
		return nil, errors.New("an-error")
	}
	return entities.EntityInterface(&entities.GenericInfrastructureEntity{Type: correctEntityType}), nil
}

func (c clientMock) FindEntityMetrics(sample, customTagKey, entityTag string) ([]string, error) {
	return []string{"powerdns_authoritative_deferred_cache_actions"}, nil
}

func (c clientMock) NRQLQuery(query, customTagKey, entityTag string) error {
	if query == errNRQLQuery {
		return errors.New("an-error")
	}
	return nil
}
