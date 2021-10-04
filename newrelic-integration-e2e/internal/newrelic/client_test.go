package newrelic

import (
	"errors"
	"fmt"
	"testing"

	"github.com/newrelic/newrelic-client-go/pkg/entities"
	"github.com/newrelic/newrelic-client-go/pkg/nrdb"
)

const (
	entityGUIDA           = "Mjc2Mjk0NXxJTkZSQXxOQXwtMzAzMjA2ODg0MjM5NDA1Nzg1OQ"
	entityGUIDB           = "Axz2Mjk0NXxJTkZSQXxOQXwtMzAzMjA2ODg0MjM5NDA1Nzg1OQ"
	sample                = "Metric"
	customTagKey          = "testKey"
	entityTag             = "uuuuxxx"
	errorMetricName       = "error-metric"
	emptyMetricName       = "empty-metric"
	withoutGUIDMetricName = "without-guid-metric"
)

var randomError = errors.New("a-random-query-error")

type apiClientMock struct{}

func (a apiClientMock) Query(_ int, query string) (*nrdb.NRDBResultContainer, error) {
	errorQuery := fmt.Sprintf(
		"SELECT * from %s where metricName = '%s' where %s = '%s' limit 1",
		sample, errorMetricName, customTagKey, entityTag,
	)
	emptyQuery := fmt.Sprintf(
		"SELECT * from %s where metricName = '%s' where %s = '%s' limit 1",
		sample, emptyMetricName, customTagKey, entityTag,
	)
	withoutGUIDQuery := fmt.Sprintf(
		"SELECT * from %s where metricName = '%s' where %s = '%s' limit 1",
		sample, withoutGUIDMetricName, customTagKey, entityTag,
	)

	switch query {
	case errorQuery:
		return nil, randomError
	case emptyQuery:
		return &nrdb.NRDBResultContainer{
			Results: nil,
		}, nil
	case withoutGUIDQuery:
		return &nrdb.NRDBResultContainer{
			Results: []nrdb.NRDBResult{
				map[string]interface{}{
					"newrelic.agentVersion": "1.20.2",
					"testKey":               "gyzsteszda",
				},
			},
		}, nil
	}

	return &nrdb.NRDBResultContainer{
		Results: []nrdb.NRDBResult{
			map[string]interface{}{
				"newrelic.agentVersion": "1.20.2",
				"entity.guid":           entityGUIDA,
				"testKey":               "gyzsteszda",
			},
		},
	}, nil
}

func (a apiClientMock) GetEntity(guid *entities.EntityGUID) (*entities.EntityInterface, error) {
	uncorrectEntity := entities.EntityGUID(fmt.Sprintf("%+v", entityGUIDA))
	nilEntity := entities.EntityGUID(fmt.Sprintf("%+v", entityGUIDB))
	switch *guid {
	case uncorrectEntity:
		return nil, randomError
	case nilEntity:
		return nil, nil
	}

	entity := entities.EntityInterface(&entities.GenericInfrastructureEntity{})
	return &entity, nil
}

func TestNrClient_FindEntityGUID(t *testing.T) {
	correctEntity := entities.EntityGUID(fmt.Sprintf("%+v", entityGUIDA))

	tests := []struct {
		name          string
		metricName    string
		entityGUID    *entities.EntityGUID
		errorExpected error
	}{
		{
			name:          "when the client call returns an error it should return it",
			metricName:    errorMetricName,
			errorExpected: randomError,
		},
		{
			name:          "when the client returns no results it should return ErrNoResult",
			metricName:    emptyMetricName,
			errorExpected: ErrNoResult,
		},
		{
			name:          "when the client returns an entity without guid it should return ErrNilEntity",
			metricName:    withoutGUIDMetricName,
			errorExpected: ErrNilGUID,
		},
		{
			name:       "when the client returns an entity with guid it should return it",
			metricName: "random-existing-metric-name",
			entityGUID: &correctEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrClient := nrClient{
				client: apiClientMock{},
			}
			guid, err := nrClient.FindEntityGUID(sample, tt.metricName, customTagKey, entityTag)
			if !errors.Is(err, tt.errorExpected) {
				t.Errorf("Error returned is not: %w", tt.errorExpected)
			}
			if guid != nil && *guid != *tt.entityGUID {
				t.Errorf("Expected: %v, got: %v", *tt.entityGUID, *guid)
			}
		})
	}
}

func TestNrClient_FindEntityByGUID(t *testing.T) {
	unCorrectEntity := entities.EntityGUID(fmt.Sprintf("%+v", entityGUIDA))
	nilEntity := entities.EntityGUID(fmt.Sprintf("%+v", entityGUIDB))
	someRandomCorrectEntity := entities.EntityGUID(fmt.Sprintf("%+v", "a-guid"))

	tests := []struct {
		name          string
		entityGUID    *entities.EntityGUID
		errorExpected error
	}{
		{
			name:          "when the GUID is nil it should return ErrNilGUID",
			entityGUID:    nil,
			errorExpected: ErrNilGUID,
		},
		{
			name:          "when the client call returns an error it should return it",
			entityGUID:    &unCorrectEntity,
			errorExpected: randomError,
		},
		{
			name:          "when the client returns a nil entity it should return ErrNilEntity",
			entityGUID:    &nilEntity,
			errorExpected: ErrNilEntity,
		},
		{
			name:          "when the client returns a correct entity it should return it",
			entityGUID:    &someRandomCorrectEntity,
			errorExpected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrClient := nrClient{
				client: apiClientMock{},
			}
			guid, err := nrClient.FindEntityByGUID(tt.entityGUID)
			if !errors.Is(err, tt.errorExpected) {
				t.Errorf("Error returned is not: %w", tt.errorExpected)
			}
			if tt.errorExpected == nil && guid == nil {
				t.Errorf("Expected entity, got nil")
			}
		})
	}
}
