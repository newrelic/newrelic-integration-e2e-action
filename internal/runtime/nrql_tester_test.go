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

func Test_validateNRQLTestConfig(t *testing.T) {
	lowerResult := 5.0
	upperResult := 15.0

	type args struct {
		nrqlTest spec.TestNRQL
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "a backwards compatible test without expectedError does not return an error",
			args: args{nrqlTest: spec.TestNRQL{
				Query:           "FROM Metric SELECT sum(k8s.container.restartCountDelta)",
				ErrorExpected:   false,
				ExpectedResults: nil,
			}},
			wantErr: false,
		},
		{
			name: "a backwards compatible test with expectedError does not return an error",
			args: args{nrqlTest: spec.TestNRQL{
				Query:           "FROM Metric SELECT sum(k8s.container.restartCountDelta)",
				ErrorExpected:   true,
				ExpectedResults: nil,
			}},
			wantErr: false,
		},
		{
			name: "an expected_results test with error_expected true returns an error",
			args: args{nrqlTest: spec.TestNRQL{
				Query:         "FROM Metric SELECT sum(k8s.container.restartCountDelta)",
				ErrorExpected: true,
				ExpectedResults: []spec.TestNRQLExpectedResult{
					{
						Key:               "restartCountDelta",
						Value:             123,
						LowerBoundedValue: nil,
						UpperBoundedValue: nil,
					},
				},
			}},
			wantErr: true,
		},
		{
			name: "an expected_results test with value and bounds returns an error",
			args: args{nrqlTest: spec.TestNRQL{
				Query:         "FROM Metric SELECT sum(k8s.container.restartCountDelta)",
				ErrorExpected: true,
				ExpectedResults: []spec.TestNRQLExpectedResult{
					{
						Key:               "restartCountDelta",
						Value:             123,
						LowerBoundedValue: &lowerResult,
						UpperBoundedValue: &upperResult,
					},
				},
			}},
			wantErr: true,
		},
		{
			name: "an expected_results test with no value and no bounds returns an error",
			args: args{nrqlTest: spec.TestNRQL{
				Query:         "FROM Metric SELECT sum(k8s.container.restartCountDelta)",
				ErrorExpected: true,
				ExpectedResults: []spec.TestNRQLExpectedResult{
					{
						Key:               "restartCountDelta",
						Value:             nil,
						LowerBoundedValue: nil,
						UpperBoundedValue: nil,
					},
				},
			}},
			wantErr: true,
		},
		{
			name: "a valid expected_results test returns no error",
			args: args{nrqlTest: spec.TestNRQL{
				Query:         "FROM Metric SELECT sum(k8s.container.restartCountDelta)",
				ErrorExpected: true,
				ExpectedResults: []spec.TestNRQLExpectedResult{
					{
						Key:               "restartCountDelta",
						Value:             123,
						LowerBoundedValue: nil,
						UpperBoundedValue: nil,
					},
					{
						Key:               "restartCountDelta",
						Value:             nil,
						LowerBoundedValue: &lowerResult,
						UpperBoundedValue: nil,
					},
					{
						Key:               "restartCountDelta",
						Value:             nil,
						LowerBoundedValue: nil,
						UpperBoundedValue: &upperResult,
					},
					{
						Key:               "restartCountDelta",
						Value:             nil,
						LowerBoundedValue: &lowerResult,
						UpperBoundedValue: &upperResult,
					},
				},
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateNRQLTestConfig(tt.args.nrqlTest); (err != nil) != tt.wantErr {
				t.Errorf("validateNRQLTestConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
