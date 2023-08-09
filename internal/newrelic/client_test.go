package newrelic

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-client-go/pkg/common"
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
		"SELECT uniques(entity.guid) from %s where metricName = '%s' where %s = '%s' limit 1",
		sample, errorMetricName, customTagKey, entityTag,
	)
	emptyQuery := fmt.Sprintf(
		"SELECT uniques(entity.guid) from %s where metricName = '%s' where %s = '%s' limit 1",
		sample, emptyMetricName, customTagKey, entityTag,
	)
	withoutGUIDQuery := fmt.Sprintf(
		"SELECT uniques(entity.guid) from %s where metricName = '%s' where %s = '%s' limit 1",
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
				"uniques.entity.guid":   []interface{}{entityGUIDA, entityGUIDB},
				"testKey":               "gyzsteszda",
			},
		},
	}, nil
}

func (a apiClientMock) GetEntity(guid *common.EntityGUID) (*entities.EntityInterface, error) {
	uncorrectEntity := common.EntityGUID(fmt.Sprintf("%+v", entityGUIDA))
	nilEntity := common.EntityGUID(fmt.Sprintf("%+v", entityGUIDB))
	switch *guid {
	case uncorrectEntity:
		return nil, randomError
	case nilEntity:
		return nil, nil
	}

	entity := entities.EntityInterface(&entities.GenericInfrastructureEntity{})
	return &entity, nil
}

func TestNrClient_FindEntityGUIDs(t *testing.T) {
	correctEntityA := common.EntityGUID(fmt.Sprintf("%+v", entityGUIDA))
	correctEntityB := common.EntityGUID(fmt.Sprintf("%+v", entityGUIDB))

	tests := []struct {
		name           string
		metricName     string
		entityGUIDs    []common.EntityGUID
		expectedNumber int
		errorExpected  error
	}{
		{
			name:           "when the client call returns an error it should return it",
			metricName:     errorMetricName,
			expectedNumber: 2,
			errorExpected:  randomError,
		},
		{
			name:           "when the client returns no results it should return ErrNoResult",
			metricName:     emptyMetricName,
			expectedNumber: 2,
			errorExpected:  ErrNoResult,
		},
		{
			name:           "when the client returns no results it should return ErrNoResult",
			metricName:     emptyMetricName,
			expectedNumber: 2,
			errorExpected:  ErrNoResult,
		},
		{
			name:           "when the client returns uniques.entity.guid smaller than expected number it should return ErrResultNumber",
			metricName:     "random-existing-metric-name",
			expectedNumber: 3,
			errorExpected:  ErrResultNumber,
		},
		{
			name:           "when the client returns uniques.entity.guid equal to expected it should return the array",
			metricName:     "random-existing-metric-name",
			expectedNumber: 2,
			entityGUIDs:    []common.EntityGUID{correctEntityA, correctEntityB},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrClient := nrClient{
				client: apiClientMock{},
			}
			guid, err := nrClient.FindEntityGUIDs(sample, tt.metricName, customTagKey, entityTag, tt.expectedNumber)
			if !errors.Is(err, tt.errorExpected) {
				t.Errorf("Error expected: %v, error returned: %v", tt.errorExpected, err)
			}
			if guid != nil && !reflect.DeepEqual(guid, tt.entityGUIDs) {
				t.Errorf("Expected: %v, got: %v", tt.entityGUIDs, tt.entityGUIDs)
			}
		})
	}
}

func TestNrClient_FindEntityByGUID(t *testing.T) {
	unCorrectEntity := common.EntityGUID(fmt.Sprintf("%+v", entityGUIDA))
	nilEntity := common.EntityGUID(fmt.Sprintf("%+v", entityGUIDB))
	someRandomCorrectEntity := common.EntityGUID(fmt.Sprintf("%+v", "a-guid"))

	tests := []struct {
		name          string
		entityGUID    *common.EntityGUID
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
				t.Errorf("Error returned is not: %v", tt.errorExpected)
			}
			if tt.errorExpected == nil && guid == nil {
				t.Errorf("Expected entity, got nil")
			}
		})
	}
}

func Test_checkBounds(t *testing.T) {
	lowerResult := 5.0
	upperResult := 15.0

	type args struct {
		actualResult        any
		expectedLowerResult *float64
		expectedUpperResult *float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "when the actual result is below of a dual-bound it should return an error",
			args:    args{actualResult: 2.0, expectedLowerResult: &lowerResult, expectedUpperResult: &upperResult},
			wantErr: true,
		},
		{
			name:    "when the actual result is above of a dual-bound it should return an error",
			args:    args{actualResult: 20.0, expectedLowerResult: &lowerResult, expectedUpperResult: &upperResult},
			wantErr: true,
		},
		{
			name:    "when the actual result is below of an lower bound it should return an error",
			args:    args{actualResult: 1.0, expectedLowerResult: &lowerResult, expectedUpperResult: nil},
			wantErr: true,
		},
		{
			name:    "when the actual result is above of an upper bound it should return an error",
			args:    args{actualResult: 16.0, expectedLowerResult: nil, expectedUpperResult: &upperResult},
			wantErr: true,
		},
		{
			name:    "when the actual result is above of an lower bound it should return no error",
			args:    args{actualResult: 6.0, expectedLowerResult: &lowerResult, expectedUpperResult: nil},
			wantErr: false,
		},
		{
			name:    "when the actual result is equal to a lower bound it should return no error",
			args:    args{actualResult: 5.0, expectedLowerResult: &lowerResult, expectedUpperResult: nil},
			wantErr: false,
		},
		{
			name:    "when the actual result is below of an upper bound it should return no error",
			args:    args{actualResult: 14.0, expectedLowerResult: nil, expectedUpperResult: &upperResult},
			wantErr: false,
		},
		{
			name:    "when the actual result is equal to an upper bound it should return no error",
			args:    args{actualResult: 15.0, expectedLowerResult: nil, expectedUpperResult: &upperResult},
			wantErr: false,
		},
		{
			name:    "when the actual result is between a dual-bound it should return no error",
			args:    args{actualResult: 10.0, expectedLowerResult: &lowerResult, expectedUpperResult: &upperResult},
			wantErr: false,
		},
		{
			name:    "when the actual result is nil an error should be thrown",
			args:    args{actualResult: nil, expectedLowerResult: nil, expectedUpperResult: nil},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkBounds(tt.args.actualResult, tt.args.expectedLowerResult, tt.args.expectedUpperResult); (err != nil) != tt.wantErr {
				t.Errorf("checkBounds() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
