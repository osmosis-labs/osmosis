package osmoutils_test

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

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

func TestContains(t *testing.T) {
	testCases := []struct {
		name   string
		slice  []int
		item   int
		expect bool
	}{
		{
			name:   "Contains - item is in the slice",
			slice:  []int{1, 2, 3, 4, 5},
			item:   3,
			expect: true,
		},
		{
			name:   "Contains - item is not in the slice",
			slice:  []int{1, 2, 3, 4, 5},
			item:   6,
			expect: false,
		},
		// add more test cases here...
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := osmoutils.Contains(tc.slice, tc.item)
			if got != tc.expect {
				t.Fatalf("Contains(%v, %v): expected %v, got %v", tc.slice, tc.item, tc.expect, got)
			}
		})
	}
}

func TestGetRandomSubset(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
	}{
		{
			name:  "Empty slice",
			slice: []int{},
		},
		{
			name:  "Slice of integers",
			slice: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
	}

	rand.Seed(time.Now().UnixNano())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := osmoutils.GetRandomSubset(tt.slice)

			// Check if the length of the returned subset is less than or equal to the length of the original slice
			if len(got) > len(tt.slice) {
				t.Errorf("GetRandomSubset() returned subset length %d, expected less than or equal to %d", len(got), len(tt.slice))
			}

			// Check if the returned subset contains only elements from the original slice
			for _, v := range got {
				if !contains(tt.slice, v) {
					t.Errorf("GetRandomSubset() returned element %v not found in the original slice", v)
				}
			}
		})
	}
}

// contains checks if a slice contains a specific element
func contains(slice interface{}, element interface{}) bool {
	s := reflect.ValueOf(slice)
	for i := 0; i < s.Len(); i++ {
		if reflect.DeepEqual(s.Index(i).Interface(), element) {
			return true
		}
	}
	return false
}
