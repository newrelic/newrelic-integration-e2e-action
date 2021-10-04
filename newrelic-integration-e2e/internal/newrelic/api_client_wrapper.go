package newrelic

import (
	newrelicgo "github.com/newrelic/newrelic-client-go/newrelic"
	"github.com/newrelic/newrelic-client-go/pkg/entities"
	"github.com/newrelic/newrelic-client-go/pkg/nrdb"
)

type ApiClient interface {
	Query(accountId int, query string) (*nrdb.NRDBResultContainer, error)
	GetEntity(guid *entities.EntityGUID) (*entities.EntityInterface, error)
}

type ApiClientWrapper struct {
	client *newrelicgo.NewRelic
}

func NewApiClientWrapper(apiKey string) (ApiClientWrapper, error) {
	client, err := newrelicgo.New(newrelicgo.ConfigPersonalAPIKey(apiKey))
	if err != nil {
		return ApiClientWrapper{}, err
	}
	return ApiClientWrapper{client: client}, err
}

func (a ApiClientWrapper) Query(accountId int, query string) (*nrdb.NRDBResultContainer, error) {
	return a.client.Nrdb.Query(accountId, nrdb.NRQL(query))
}

func (a ApiClientWrapper) GetEntity(guid *entities.EntityGUID) (*entities.EntityInterface, error) {
	return a.client.Entities.GetEntity(*guid)
}
