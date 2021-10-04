package runtime

import (
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestEntitiesTester_Test(t *testing.T) {
	log := logrus.New()
	log.SetOutput(ioutil.Discard)
	entitiesTester := NewEntitiesTester(clientMock{}, log)

	inputTests := spec.Tests{Entities: []spec.TestEntity{
		{
			Type:       "",
			DataType:   errFindEntityGUID,
			MetricName: "",
		},
		{
			Type:       "",
			DataType:   errFindEntityByGUID,
			MetricName: "",
		},
		{
			Type:       "incorrectEntityType",
			DataType:   "",
			MetricName: "",
		},
		{
			Type:       correctEntityType,
			DataType:   "",
			MetricName: "",
		},
	}}

	errors := entitiesTester.Test(inputTests, "", "")
	assert.Equal(t, 3, len(errors))
}
