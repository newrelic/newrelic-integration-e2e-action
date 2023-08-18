package newrelic

import (
	"errors"
	"fmt"
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
	ErrNilEntity        = errors.New("nil entity, impossible to dereference")
	ErrNilGUID          = errors.New("GUID is nil, impossible to find entity")
	ErrExpected         = errors.New("an error was expected")
	ErrTypeAssertion    = errors.New("could not assert type from any")
	ErrAssertionFailure = errors.New("assertion failure")
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

	a, err := nrc.client.Query(nrc.accountID, query)
	if err != nil {
		return fmt.Errorf("executing nrql query %s, %w", query, err)
	}

	if expectedResults == nil {
		// Backwards compatible test
		err := nrqlQueryDefaultTest(a.Results)
		if err != nil && !errorExpected {
			return fmt.Errorf("querying: %w: %s", err, query)
		}
		if err == nil && errorExpected {
			return fmt.Errorf("running %q: %w", query, ErrExpected)
		}
		return nil
	}

	// Expected value test
	testErr := nrqlQueryExpectedValueTest(a.Results, expectedResults)
	if testErr != nil {
		return fmt.Errorf("%w: %s", testErr, query)
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
