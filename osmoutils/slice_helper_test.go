package osmoutils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReverseSlice(t *testing.T) {
	tests := map[string]struct {
		s []string

		expectedSolvedInput []string
	}{
		"Even length array":       {[]string{"a", "b", "c", "d"}, []string{"d", "c", "b", "a"}},
		"Empty array":             {[]string{}, []string{}},
		"Odd length array":        {[]string{"a", "b", "c"}, []string{"c", "b", "a"}},
		"Single element array":    {[]string{"a"}, []string{"a"}},
		"Array with empty string": {[]string{"a", "b", "c", "", "d"}, []string{"d", "", "c", "b", "a"}},
		"Array with numbers":      {[]string{"a", "b", "c", "1", "2", "3"}, []string{"3", "2", "1", "c", "b", "a"}},
	}

	for _, tc := range tests {
		actualSolvedInput := ReverseSlice(tc.s)
		require.True(t, reflect.DeepEqual(actualSolvedInput, tc.expectedSolvedInput))
	}
}
