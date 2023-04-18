package osmoutils_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmoutils"
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

func TestMergeSlices(t *testing.T) {
	lessInt := func(a, b int) bool {
		return a < b
	}
	testCases := []struct {
		name   string
		slice1 []int
		slice2 []int
		less   func(a, b int) bool
		want   []int
	}{
		{
			name:   "basic merge",
			slice1: []int{1, 3, 5},
			slice2: []int{2, 4, 6},
			less:   lessInt,
			want:   []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:   "Empty slice1",
			slice1: []int{},
			slice2: []int{2, 4, 6},
			less:   lessInt,
			want:   []int{2, 4, 6},
		},
		{
			name:   "Empty slice2",
			slice1: []int{1, 3, 5},
			slice2: []int{},
			less:   lessInt,
			want:   []int{1, 3, 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := osmoutils.MergeSlices(tc.slice1, tc.slice2, lessInt)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got: %v, want: %v", got, tc.want)
			}
		})
	}
}

func TestContainsDuplicateDeepEqual(t *testing.T) {
	tests := []struct {
		input []interface{}
		want  bool
	}{
		{[]interface{}{[]int{1, 2, 3}, []int{4, 5, 6}}, false},
		{[]interface{}{[]int{1, 2, 3}, []int{1, 2, 3}}, true},
		{[]interface{}{[]string{"hello", "world"}, []string{"goodbye", "world"}}, false},
		{[]interface{}{[]string{"hello", "world"}, []string{"hello", "world"}}, true},
		{[]interface{}{[][]int{{1, 2}, {3, 4}}, [][]int{{1, 2}, {3, 4}}}, true},
	}

	for _, tt := range tests {
		got := osmoutils.ContainsDuplicateDeepEqual(tt.input)
		require.Equal(t, tt.want, got)
	}
}
