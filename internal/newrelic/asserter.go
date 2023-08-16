package newrelic

import (
	"errors"
	"fmt"
	"strings"

	"github.com/newrelic/newrelic-client-go/pkg/nrdb"
	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"
)

var (
	ErrNoResult          = errors.New("query did not return any result")
	ErrNotValid          = errors.New("query did not return a valid result")
	ErrResultNumber      = errors.New("query did not return expected number of results")
	ErrNotExpectedResult = errors.New("query did not return expected results")
)

func nrqlQueryDefaultTest(queryResults []nrdb.NRDBResult) error {
	if len(queryResults) == 0 {
		return ErrNoResult
	}
	if !validValue(queryResults) {
		return ErrNotValid
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

func nrqlQueryExpectedValueTest(queryResults []nrdb.NRDBResult, expectedResults []spec.TestNRQLExpectedResult) error {
	if len(expectedResults) != len(queryResults) {
		return fmt.Errorf("%w: expected %d got %d", ErrResultNumber, len(expectedResults), len(queryResults))
	}
	for i, expectedResult := range expectedResults {
		actualResult := queryResults[i][expectedResult.Key]
		comparisonErr := compareResults(actualResult, expectedResult)
		if comparisonErr != nil {
			return fmt.Errorf("%w: for key '%s': %s", ErrNotExpectedResult, expectedResult.Key, comparisonErr.Error())
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
	actualFloat, err := extractFloat(actualResult)
	if err != nil {
		return err
	}
	return checkBounds(actualFloat, expectedResult.LowerBoundedValue, expectedResult.UpperBoundedValue)
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

func checkBounds(actualFloat float64, expectedLowerResult *float64, expectedUpperResult *float64) error {
	if (expectedLowerResult == nil || *expectedLowerResult <= actualFloat) && (expectedUpperResult == nil || actualFloat <= *expectedUpperResult) {
		return nil
	}
	rangeAsString := formatRange(expectedLowerResult, expectedUpperResult)
	return fmt.Errorf("%w - expected value in range %s, got %f", ErrAssertionFailure, rangeAsString, actualFloat)
}

// formatRange returns "[-Inf, x]" "[x, Inf]" or "[x, y]" depending of bounds
func formatRange(lowerBound *float64, upperBound *float64) string {
	if lowerBound == nil {
		return fmt.Sprintf("[-INF,%f]", *upperBound)
	} else if upperBound == nil {
		return fmt.Sprintf("[%f,INF]", *lowerBound)
	}
	return fmt.Sprintf("[%f,%f]", *lowerBound, *upperBound)
}
