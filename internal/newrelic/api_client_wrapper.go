package newrelic

import (
	"fmt"
	newrelicgo "github.com/newrelic/newrelic-client-go/newrelic"
	"github.com/newrelic/newrelic-client-go/pkg/common"
	"github.com/newrelic/newrelic-client-go/pkg/entities"
	"github.com/newrelic/newrelic-client-go/pkg/nrdb"
	"github.com/newrelic/newrelic-client-go/pkg/region"
)

type ApiClient interface {
	Query(accountId int, query string) (*nrdb.NRDBResultContainer, error)
	GetEntity(guid *common.EntityGUID) (*entities.EntityInterface, error)
}

type ApiClientWrapper struct {
	client *newrelicgo.NewRelic
}

func NewApiClientWrapper(apiKey string, apiRegion string) (ApiClientWrapper, error) {
	if _, ok := region.Regions[region.Name(apiRegion)]; !ok {
		return ApiClientWrapper{}, fmt.Errorf("region %s is not valid", apiRegion)
	}

	client, err := newrelicgo.New(newrelicgo.ConfigPersonalAPIKey(apiKey), newrelicgo.ConfigRegion(apiRegion))
	if err != nil {
		return ApiClientWrapper{}, err
	}
	return ApiClientWrapper{client: client}, err
}

func (a ApiClientWrapper) Query(accountId int, query string) (*nrdb.NRDBResultContainer, error) {
	return a.client.Nrdb.Query(accountId, nrdb.NRQL(query))
}

func (a ApiClientWrapper) GetEntity(guid *common.EntityGUID) (*entities.EntityInterface, error) {
	return a.client.Entities.GetEntity(*guid)
}
