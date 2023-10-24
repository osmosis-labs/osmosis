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

func TestDisjointArrays(t *testing.T) {
	tests := []struct {
		name string
		a    []uint64
		b    []uint64
		want []uint64
	}{
		{
			name: "Both arrays contain some unique elements, both sorted",
			a:    []uint64{1, 2, 3, 4, 5},
			b:    []uint64{4, 5, 6, 7, 8},
			want: []uint64{1, 2, 3, 6, 7, 8},
		},
		{
			name: "Only array 'a' contains some unique elements, both sorted",
			a:    []uint64{10, 20, 30, 40},
			b:    []uint64{20, 30},
			want: []uint64{10, 40},
		},
		{
			name: "Only array 'b' contains some unique elements, both sorted",
			a:    []uint64{20, 30},
			b:    []uint64{10, 20, 30, 40},
			want: []uint64{10, 40},
		},
		{
			name: "Both arrays contain some unique elements, both unsorted",
			a:    []uint64{5, 4, 3, 2, 1},
			b:    []uint64{8, 7, 6, 5, 4},
			want: []uint64{1, 2, 3, 6, 7, 8},
		},
		{
			name: "Both arrays contain some unique elements, a sorted, b unsorted",
			a:    []uint64{5, 4, 3, 2, 1},
			b:    []uint64{6, 7, 8, 5, 4},
			want: []uint64{1, 2, 3, 6, 7, 8},
		},
		{
			name: "Only array 'a' contains some unique elements, both unsorted",
			a:    []uint64{40, 30, 20, 10},
			b:    []uint64{30, 20},
			want: []uint64{10, 40},
		},
		{
			name: "Only array 'b' contains some unique elements, both unsorted",
			a:    []uint64{30, 20},
			b:    []uint64{40, 30, 20, 10},
			want: []uint64{10, 40},
		},
		{
			name: "Both arrays empty",
			a:    []uint64{},
			b:    []uint64{},
			want: []uint64{},
		},
		{
			name: "Array 'a' empty, unsorted",
			a:    []uint64{},
			b:    []uint64{2, 1, 3},
			want: []uint64{1, 2, 3},
		},
		{
			name: "Array 'b' empty, unsorted",
			a:    []uint64{2, 1, 3},
			b:    []uint64{},
			want: []uint64{1, 2, 3},
		},
		{
			name: "Both arrays contain completely different elements, both unsorted",
			a:    []uint64{2, 1, 3},
			b:    []uint64{5, 4, 6},
			want: []uint64{1, 2, 3, 4, 5, 6},
		},
		{
			name: "Both arrays equal, both unsorted",
			a:    []uint64{2, 1, 3},
			b:    []uint64{2, 1, 3},
			want: []uint64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := osmoutils.DisjointArrays(tt.a, tt.b)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DifferenceBetweenUint64Arrays() = %v, want %v", got, tt.want)
			}
		})
	}
}
