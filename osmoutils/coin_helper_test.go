package osmoutils_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

func TestSubDecCoins(t *testing.T) {
	var (
	emptyCoins      = sdk.DecCoins(nil)
	fiftyFooCoins = sdk.NewDecCoin("foo", sdk.NewInt(50))
	fiftyBarCoins = sdk.NewDecCoin("bar", sdk.NewInt(50))
	hundredFooCoins = sdk.NewDecCoin("foo", sdk.NewInt(100))
	hundredBarCoins = sdk.NewDecCoin("bar", sdk.NewInt(100))

	fiftyEach = sdk.NewDecCoins(fiftyFooCoins, fiftyBarCoins)
	hundredEach = sdk.NewDecCoins(hundredFooCoins, hundredBarCoins)
	)

	tests := map[string]struct {
		firstInput []sdk.DecCoins
		secondInput []sdk.DecCoins

		expectedOutput []sdk.DecCoins
		expectError bool
	}{
		"[[100foo, 100bar], [100foo, 100bar]] - [[50foo, 50bar], [50foo, 100bar]]": {
			firstInput: []sdk.DecCoins{hundredEach, hundredEach},
			secondInput: []sdk.DecCoins{fiftyEach, hundredEach},

			expectedOutput: []sdk.DecCoins{fiftyEach, emptyCoins},
		},
		"[[100bar, 100foo], [100foo, 100bar]] - [[50foo, 50bar], [50foo, 100bar]]": {
			firstInput: []sdk.DecCoins{
				sdk.NewDecCoins(hundredBarCoins, hundredFooCoins), 
				hundredEach},
			secondInput: []sdk.DecCoins{fiftyEach, hundredEach},

			expectedOutput: []sdk.DecCoins{fiftyEach, emptyCoins},
		},
		"both inputs empty": {
			firstInput: []sdk.DecCoins{},
			secondInput: []sdk.DecCoins{},

			expectedOutput: []sdk.DecCoins{},
		},
		"[[100foo]] - [[50foo]]": {
			firstInput: []sdk.DecCoins{sdk.NewDecCoins(hundredFooCoins)},
			secondInput: []sdk.DecCoins{sdk.NewDecCoins(fiftyFooCoins)},

			expectedOutput: []sdk.DecCoins{sdk.NewDecCoins(fiftyFooCoins)},
		},

		// error catching

		"different length inputs": {
			firstInput: []sdk.DecCoins{hundredEach, hundredEach, hundredEach},
			secondInput: []sdk.DecCoins{fiftyEach, hundredEach},

			expectedOutput: []sdk.DecCoins{},
			expectError: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualOutput, err := osmoutils.SubDecCoinArrays(tc.firstInput, tc.secondInput)

			if tc.expectError {
				require.Error(t, err)
				require.Equal(t, tc.expectedOutput, actualOutput)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedOutput, actualOutput)
		})
	}
}
