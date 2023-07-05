package osmoutils_test

import (
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
