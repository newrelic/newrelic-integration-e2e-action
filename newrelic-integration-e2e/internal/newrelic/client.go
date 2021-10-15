package newrelic

import (
	"errors"
	"fmt"
	"log"

	"github.com/newrelic/newrelic-client-go/pkg/nrdb"

	"github.com/newrelic/newrelic-client-go/pkg/entities"
)

type Client interface {
	FindEntityGUIDs(sample, metricName, customTagKey, entityTag string, expectedNumber int) ([]entities.EntityGUID, error)
	FindEntityByGUID(guid *entities.EntityGUID) (entities.EntityInterface, error)
	FindEntityMetrics(sample, customTagKey, entityTag string) ([]string, error)
	NRQLQuery(query, customTagKey, entityTag string) error
}

var (
	ErrNilEntity    = errors.New("nil entity, impossible to dereference")
	ErrNilGUID      = errors.New("GUID is nil, impossible to find entity")
	ErrNoResult     = errors.New("query did not return any result")
	ErrResultNumber = errors.New("query did not return expected number of results")
	ErrNotValid     = errors.New("query did not return a valid result")
)

type nrClient struct {
	accountID int
	apiKey    string
	client    ApiClient
}

func NewNrClient(apiKey string, accountID int) *nrClient {
	client, err := NewApiClientWrapper(apiKey)
	if err != nil {
		log.Fatal("error initializing client:", err)
	}
	return &nrClient{
		client:    client,
		apiKey:    apiKey,
		accountID: accountID,
	}
}

func (nrc *nrClient) FindEntityGUIDs(sample, metricName, customTagKey, entityTag string, expectedNumber int) ([]entities.EntityGUID, error) {
	var entityGuids []entities.EntityGUID
	query := fmt.Sprintf("SELECT uniques(entity.guid) from %s where metricName = '%s' where %s = '%s' limit 1", sample, metricName, customTagKey, entityTag)

	a, err := nrc.client.Query(nrc.accountID, query)
	if err != nil {
		return nil, fmt.Errorf("executing query to fetch entity GUIDs %s, %w", query, err)
	}

	if a.Results[0]["uniques.entity.guid"] == nil {
		return nil, ErrNoResult
	}

	if len(a.Results[0]["uniques.entity.guid"].([]interface{})) < expectedNumber {
		return nil, fmt.Errorf("%w: %s", ErrResultNumber, query)
	}

	for _, g := range a.Results[0]["uniques.entity.guid"].([]interface{}) {
		guid := entities.EntityGUID(fmt.Sprintf("%v", g))
		entityGuids = append(entityGuids, guid)
	}

	return entityGuids, nil
}

func (nrc *nrClient) FindEntityByGUID(guid *entities.EntityGUID) (entities.EntityInterface, error) {
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

func (nrc *nrClient) NRQLQuery(query, customTagKey, entityTag string) error {
	query = fmt.Sprintf("%s WHERE %s = '%s'", query, customTagKey, entityTag)

	a, err := nrc.client.Query(nrc.accountID, query)
	if err != nil {
		return fmt.Errorf("executing nrql query %s, %w", query, err)
	}
	if len(a.Results) == 0 {
		return fmt.Errorf("%w: %s", ErrNoResult, query)
	}
	if !validValue(a.Results) {
		return fmt.Errorf("%w: %s", ErrNotValid, query)
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
