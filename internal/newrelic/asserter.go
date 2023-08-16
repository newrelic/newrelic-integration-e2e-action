package newrelic

import (
	"errors"
	"fmt"
	"github.com/newrelic/newrelic-client-go/pkg/nrdb"
	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"
	"strings"
)

var (
	ErrNoResult          = errors.New("query did not return any result")
	ErrNotValid          = errors.New("query did not return a valid result")
	ErrResultNumber      = errors.New("query did not return expected number of results")
	ErrNotExpectedResult = errors.New("query did not return expected results")
)

func nrqlQueryDefaultTest(nrc *nrClient, query string) error {
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

func nrqlQueryExpectedValueTest(nrc *nrClient, query string, expectedResults []spec.TestNRQLExpectedResult) error {
	a, _ := nrc.client.Query(nrc.accountID, query)

	if len(expectedResults) != len(a.Results) {
		return fmt.Errorf("%w: %s\n - expected %d got %d", ErrResultNumber, query, len(expectedResults), len(a.Results))
	}
	for i, expectedResult := range expectedResults {
		actualResult := a.Results[i][expectedResult.Key]
		comparisonErr := compareResults(actualResult, expectedResult)
		if comparisonErr != nil {
			return fmt.Errorf("%w: %s\n - for key '%s': %s", ErrNotExpectedResult, query, expectedResult.Key, comparisonErr.Error())
		}
	}
	return nil
}

func compareResults(actualResult any, expectedResult spec.TestNRQLExpectedResult) error {
	expectedExactResult := expectedResult.Value
	if expectedExactResult != nil {
		// We are checking for an exact value
		expectedExactResult = preprocessResult(expectedExactResult)
		actualResult = preprocessResult(actualResult)

		if expectedExactResult == actualResult {
			return nil
		}
		return fmt.Errorf("%w - expected: '%s', got '%s'", ErrAssertionFailure, expectedExactResult, actualResult)
	}

	// We are checking for a bounded value
	return checkBounds(actualResult, expectedResult.LowerBoundedValue, expectedResult.UpperBoundedValue)
}

func preprocessResult(result any) any {
	switch typedResult := result.(type) {
	case int:
		intResult := typedResult
		// Convert int into floats
		return float64(intResult)
	case string:
		stringResult := typedResult
		// Convert string nil into nil
		if strings.EqualFold(stringResult, "nil") {
			return nil
		}

		// Convert string booleans into boolean type
		if strings.EqualFold(stringResult, "false") {
			return false
		} else if strings.EqualFold(stringResult, "true") {
			return true
		}
		return stringResult
	}

	return result
}

func extractFloat(result any) (float64, error) {
	result = preprocessResult(result)
	floatResult, ok := result.(float64)
	if !ok {
		return 0, fmt.Errorf("%w: float", ErrTypeAssertion)
	}
	return floatResult, nil
}

func checkBounds(actualResult any, expectedLowerResult *float64, expectedUpperResult *float64) error {
	actualFloat, err := extractFloat(actualResult)
	if err != nil {
		return err
	}

	// if either expectedLowerResult is nil or expectedLowerResult <= result, AND either expectedUpperResult is nil or result <= expectedUpperResult, return nil, ELSE, error
	if (expectedLowerResult == nil || *expectedLowerResult <= actualFloat) && (expectedUpperResult == nil || actualFloat <= *expectedUpperResult) {
		return nil
	}
	rangeAsString := formatRange(expectedLowerResult, expectedUpperResult)
	return fmt.Errorf("%w - expected value in range %s, got %f", ErrAssertionFailure, rangeAsString, actualFloat)
}

// Returns "[-Inf, x]" "[x, Inf]" or "[x, y]" depending of bounds
func formatRange(lowerBound *float64, upperBound *float64) string {
	if lowerBound == nil {
		return fmt.Sprintf("[-INF,%f]", *upperBound)
	} else if upperBound == nil {
		return fmt.Sprintf("[%f,INF]", *lowerBound)
	}
	return fmt.Sprintf("[%f,%f]", *lowerBound, *upperBound)
}
