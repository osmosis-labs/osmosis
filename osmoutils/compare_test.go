package osmoutils_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
		{
			name:   "sdk.Int",
			values: []interface{}{sdk.NewInt(1), sdk.NewInt(5), sdk.NewInt(3)},
			max:    sdk.NewInt(5),
		},
		{
			name:   "sdk.Dec",
			values: []interface{}{sdk.MustNewDecFromStr("7"), sdk.MustNewDecFromStr("5.5"), sdk.MustNewDecFromStr("3.2")},
			max:    sdk.NewDec(7),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := osmoutils.Max(tc.values...)
			assert.Equal(t, tc.max, result)
		})
	}
}
