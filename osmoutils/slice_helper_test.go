package osmoutils_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v11/osmoutils"
)

func TestReverseSlice(t *testing.T) {
	tests := map[string]struct {
		s []string

		expectedSolvedInput []string
	}{
		"Even length array":       {s: []string{"a", "b", "c", "d"}, expectedSolvedInput: []string{"d", "c", "b", "a"}},
		"Empty array":             {s: []string{}, expectedSolvedInput: []string{}},
		"Odd length array":        {s: []string{"a", "b", "c"}, expectedSolvedInput: []string{"c", "b", "a"}},
		"Single element array":    {s: []string{"a"}, expectedSolvedInput: []string{"a"}},
		"Array with empty string": {s: []string{"a", "b", "c", "", "d"}, expectedSolvedInput: []string{"d", "", "c", "b", "a"}},
		"Array with numbers":      {s: []string{"a", "b", "c", "1", "2", "3"}, expectedSolvedInput: []string{"3", "2", "1", "c", "b", "a"}},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualSolvedInput := osmoutils.ReverseSlice(tc.s)
			require.Equal(t, tc.expectedSolvedInput, actualSolvedInput)
		})
	}
}
