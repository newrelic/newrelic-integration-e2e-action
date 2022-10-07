package runtime

import (
	"io/ioutil"
	"testing"

	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
