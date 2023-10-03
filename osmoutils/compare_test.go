package osmoutils_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

func TestMax(t *testing.T) {
	testCases := []struct {
		name   string
		values []interface{}
		max    interface{}
	}{
		{
			name:   "Empty values",
			values: []interface{}{},
			max:    nil,
		},
		{
			name:   "int",
			values: []interface{}{1, 5, 3, 2},
			max:    5,
		},
		{
			name:   "uint",
			values: []interface{}{uint(10), uint(7), uint(9)},
			max:    uint(10),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := osmoutils.Max(tc.values...)
			assert.Equal(t, tc.max, result)
		})
	}
}

func TestDifferenceUint64(t *testing.T) {
	testCases := []struct {
		a        []uint64
		b        []uint64
		expected []uint64
	}{
		{
			a:        []uint64{1, 2, 3, 4, 5},
			b:        []uint64{4, 5, 6, 7, 8},
			expected: []uint64{1, 2, 3},
		},
		{
			a:        []uint64{10, 20, 30, 40, 50},
			b:        []uint64{30, 40, 50, 60, 70},
			expected: []uint64{10, 20},
		},
		{
			a:        []uint64{},
			b:        []uint64{1, 2, 3},
			expected: []uint64{},
		},
		{
			a:        []uint64{1, 2, 3},
			b:        []uint64{},
			expected: []uint64{1, 2, 3},
		},
	}

	for _, tc := range testCases {
		result := osmoutils.DifferenceBetweenUint64Arrays(tc.a, tc.b)
		if !reflect.DeepEqual(result, tc.expected) {
			t.Errorf("DifferenceUint64(%v, %v) = %v; want %v", tc.a, tc.b, result, tc.expected)
		}
	}
}
