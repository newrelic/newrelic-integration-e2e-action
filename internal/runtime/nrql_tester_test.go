package runtime

import (
	"io/ioutil"
	"testing"

	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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

func TestNRQLTester_Test_error(t *testing.T) {
	log := logrus.New()
	log.SetOutput(ioutil.Discard)
	nrqlTester := NewNRQLTester(clientMock{}, log)

	inputTests := spec.Tests{NRQLs: []spec.TestNRQL{
		{Query: errNRQLQuery, ErrorExpected: true},
		{Query: "a-correct-query"},
	}}

	errors := nrqlTester.Test(inputTests, "", "")
	assert.Equal(t, 0, len(errors))

	inputTests = spec.Tests{NRQLs: []spec.TestNRQL{
		{Query: errNRQLQuery},
		{Query: "a-correct-query", ErrorExpected: true},
	}}

	errors = nrqlTester.Test(inputTests, "", "")
	assert.Equal(t, 2, len(errors))
}
