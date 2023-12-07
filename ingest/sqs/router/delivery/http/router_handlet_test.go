package http_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/delivery/http"
)

// TestParseNumbers tests parsing a string of numbers to a slice of uint64
func TestParseNumbers(t *testing.T) {
	testCases := []struct {
		input           string
		expectedNumbers []uint64
		expectedError   bool
	}{
		{"", nil, false},                          // Empty string, expecting an empty slice and no error
		{"1,2,3", []uint64{1, 2, 3}, false},       // Comma-separated numbers, expecting slice {1, 2, 3} and no error
		{"42", []uint64{42}, false},               // Single number, expecting slice {42} and no error
		{"10,20,30", []uint64{10, 20, 30}, false}, // Another set of numbers

		// Add more test cases as needed
		{"abc", nil, true}, // Invalid input, expecting an error
	}

	for _, testCase := range testCases {
		actualNumbers, actualError := http.ParseNumbers(testCase.input)

		if testCase.expectedError {
			require.Error(t, actualError)
			return
		}

		// Check if the actual output matches the expected output
		if !reflect.DeepEqual(actualNumbers, testCase.expectedNumbers) {
			t.Errorf("Input: %s, Expected Numbers: %v, Actual Numbers: %v",
				testCase.input, testCase.expectedNumbers, actualNumbers)
		}
	}
}
