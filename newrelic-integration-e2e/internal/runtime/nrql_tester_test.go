package runtime

import (
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestNRQLTester_Test(t *testing.T) {
	log := logrus.New()
	log.SetOutput(ioutil.Discard)
	nrqlTester := NewNRQLTester(clientMock{}, log)

	inputTests := spec.Tests{NRQLs: []spec.TestNRQL{
		{Query: errNRQLQuery},
		{Query: "a-correct-query"},
	}}

	errors := nrqlTester.Test(inputTests, "", "")
	assert.Equal(t, 1, len(errors))
}
