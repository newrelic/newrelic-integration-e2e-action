package newrelic

import (
	"errors"
	"fmt"
	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"
	"log"
	"strings"

	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"

	"github.com/newrelic/newrelic-client-go/pkg/common"
	"github.com/newrelic/newrelic-client-go/pkg/entities"
	"github.com/newrelic/newrelic-client-go/pkg/nrdb"
)

type Client interface {
	FindEntityGUIDs(sample, metricName, customTagKey, entityTag string, expectedNumber int) ([]common.EntityGUID, error)
	FindEntityByGUID(guid *common.EntityGUID) (entities.EntityInterface, error)
	FindEntityMetrics(sample, customTagKey, entityTag string) ([]string, error)
	NRQLQuery(query, customTagKey, entityTag string, errorExpected bool, expectedResults []spec.TestNRQLExpectedResult) error
}

var (
	ErrNilEntity         = errors.New("nil entity, impossible to dereference")
	ErrNilGUID           = errors.New("GUID is nil, impossible to find entity")
	ErrNoResult          = errors.New("query did not return any result")
	ErrResultNumber      = errors.New("query did not return expected number of results")
	ErrNotValid          = errors.New("query did not return a valid result")
	ErrNotExpectedResult = errors.New("query did not return expected results")
	ErrExpected          = errors.New("an error was expected")
)

type nrClient struct {
	accountID int
	apiKey    string
	client    ApiClient
}

func NewNrClient(apiKey string, region string, accountID int) *nrClient {
	client, err := NewApiClientWrapper(apiKey, region)
	if err != nil {
		log.Fatal("error initializing client:", err)
	}
	return &nrClient{
		client:    client,
		apiKey:    apiKey,
		accountID: accountID,
	}
}

func (nrc *nrClient) FindEntityGUIDs(sample, metricName, customTagKey, entityTag string, expectedNumber int) ([]common.EntityGUID, error) {
	var entityGuids []common.EntityGUID
	query := fmt.Sprintf("SELECT uniques(entity.guid) from %s where metricName = '%s' where %s = '%s' limit 1", sample, metricName, customTagKey, entityTag)

	a, err := nrc.client.Query(nrc.accountID, query)
	if err != nil {
		return nil, fmt.Errorf("executing query to fetch entity GUIDs %s, %w", query, err)
	}

	if len(a.Results) < 1 || a.Results[0]["uniques.entity.guid"] == nil {
		return nil, ErrNoResult
	}

	if results := len(a.Results[0]["uniques.entity.guid"].([]interface{})); results < expectedNumber {
		return nil, fmt.Errorf("%w: %s: got %d, expected %d", ErrResultNumber, query, results, expectedNumber)
	}

	for _, g := range a.Results[0]["uniques.entity.guid"].([]interface{}) {
		guid := common.EntityGUID(fmt.Sprintf("%v", g))
		entityGuids = append(entityGuids, guid)
	}

	return entityGuids, nil
}

func (nrc *nrClient) FindEntityByGUID(guid *common.EntityGUID) (entities.EntityInterface, error) {
	if guid == nil {
		return nil, ErrNilGUID
	}

	entity, err := nrc.client.GetEntity(guid)
	if err != nil {
		return nil, fmt.Errorf("get entity: %w", err)
	}

	if entity == nil {
		return nil, ErrNilEntity
	}
	return *entity, nil
}

func (nrc *nrClient) FindEntityMetrics(sample, customTagKey, entityTag string) ([]string, error) {
	query := fmt.Sprintf("SELECT keyset() from %s where %s = '%s'", sample, customTagKey, entityTag)

	a, err := nrc.client.Query(nrc.accountID, query)
	if err != nil {
		return nil, fmt.Errorf("executing query to keyset %s, %w", query, err)
	}
	if len(a.Results) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoResult, query)
	}
	return resultMetrics(a.Results), nil
}

func (nrc *nrClient) NRQLQuery(query, customTagKey, entityTag string, errorExpected bool, expectedResults []spec.TestNRQLExpectedResult) error {
	query = fmt.Sprintf("%s WHERE %s = '%s'", query, customTagKey, entityTag)
	query = strings.ReplaceAll(query, "${SCENARIO_TAG}", entityTag)

	if expectedResults == nil {
		// Backwards compatible test
		err := nrqlQueryDefaultTest(nrc, query)
		if err != nil && !errorExpected {
			return fmt.Errorf("querying: %w", err)
		}
		if err == nil && errorExpected {
			return fmt.Errorf("running %q: %w", query, ErrExpected)
		}
		return nil
	}

	// Expected value test
	testErr := nrqlQueryExpectedValueTest(nrc, query, expectedResults)
	if testErr != nil {
		return testErr
	}
	return nil
}

func nrqlQueryDefaultTest(nrc *nrClient, query string) error {
	a, err := nrc.client.Query(nrc.accountID, query)
	if err != nil {
		if errorExpected {
			return nil
		}
		return fmt.Errorf("executing nrql query %s, %w", query, err)
	}
	if len(a.Results) == 0 {
		if errorExpected {
			return nil
		}
		return fmt.Errorf("%w: %s", ErrNoResult, query)
	}

	if expectedResults == nil {
		if !validValue(a.Results) {
			return fmt.Errorf("%w: %s", ErrNotValid, query)
		}
		return nil
	} else {
		if len(expectedResults) != len(a.Results) {
			return fmt.Errorf("%w: %s - expected %d got %d", errors.New("query did not return expected number of results"), len(expectedResults), len(a.Results))
		}
		for i, expectedResult := range expectedResults {
			stringResult := fmt.Sprintf("%v", a.Results[i][expectedResult.Key])
			if stringResult != expectedResult.Value {
				return fmt.Errorf("%w: %s - expected for key '%s': '%s' got '%s'", errors.New("query did not return expected results"), query, expectedResult.Key, expectedResult.Value, stringResult)
			}
		}
		return nil
	}
}

func nrqlQueryExpectedValueTest(nrc *nrClient, query string, expectedResults []spec.TestNRQLExpectedResult) error {
	a, _ := nrc.client.Query(nrc.accountID, query)

	if len(expectedResults) != len(a.Results) {
		return fmt.Errorf("%w: %s - expected %d got %d", ErrResultNumber, query, len(expectedResults), len(a.Results))
	}
	for i, expectedResult := range expectedResults {
		stringResult := fmt.Sprintf("%v", a.Results[i][expectedResult.Key])
		if stringResult != expectedResult.Value {
			return fmt.Errorf("%w: %s - expected for key '%s': '%s' got '%s'", ErrNotExpectedResult, query, expectedResult.Key, expectedResult.Value, stringResult)
		}
	}
	return nil
}

func resultMetrics(queryResults []nrdb.NRDBResult) []string {
	result := make([]string, len(queryResults))
	for _, r := range queryResults {
		result = append(result, fmt.Sprintf("%+v", r["key"]))
	}
	return result
}

func validValue(queryResults []nrdb.NRDBResult) bool {
	firstResult := queryResults[0]
	for key, val := range firstResult {
		if key == "timestamp" {
			continue
		}
		if val == nil {
			return false
		}
	}
	return true
}
