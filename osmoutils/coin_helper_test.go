package osmoutils_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
)

var (
	emptyCoins      = sdk.DecCoins(nil)
	fiftyFoo        = sdk.NewDecCoin("foo", osmomath.NewInt(50))
	fiftyBar        = sdk.NewDecCoin("bar", osmomath.NewInt(50))
	hundredFoo      = sdk.NewDecCoin("foo", osmomath.NewInt(100))
	hundredBar      = sdk.NewDecCoin("bar", osmomath.NewInt(100))
	hundredFiftyFoo = sdk.NewDecCoin("foo", osmomath.NewInt(150))
	hundredFiftyBar = sdk.NewDecCoin("bar", osmomath.NewInt(150))
	twoHundredFoo   = sdk.NewDecCoin("foo", osmomath.NewInt(200))
	twoHundredBar   = sdk.NewDecCoin("bar", osmomath.NewInt(200))

	fiftyEach        = sdk.NewDecCoins(fiftyFoo, fiftyBar)
	hundredEach      = sdk.NewDecCoins(hundredFoo, hundredBar)
	hundredFiftyEach = sdk.NewDecCoins(hundredFiftyFoo, hundredFiftyBar)
)

func TestSubDecCoins(t *testing.T) {
	tests := map[string]struct {
		firstInput  []sdk.DecCoins
		secondInput []sdk.DecCoins

		expectedOutput []sdk.DecCoins
		expectError    bool
		// whether unsafe subtraction should panic
		expectPanicUnsafe bool
	}{
		"[[100foo, 100bar], [100foo, 100bar]] - [[50foo, 50bar], [50foo, 100bar]]": {
			firstInput:  []sdk.DecCoins{hundredEach, hundredEach},
			secondInput: []sdk.DecCoins{fiftyEach, hundredEach},

			expectedOutput: []sdk.DecCoins{fiftyEach, emptyCoins},
		},
		"[[100bar, 100foo], [100foo, 100bar]] - [[50foo, 50bar], [50foo, 100bar]]": {
			firstInput: []sdk.DecCoins{
				sdk.NewDecCoins(hundredBar, hundredFoo),
				hundredEach},
			secondInput: []sdk.DecCoins{fiftyEach, hundredEach},

			expectedOutput: []sdk.DecCoins{fiftyEach, emptyCoins},
		},
		"both inputs empty": {
			firstInput:  []sdk.DecCoins{},
			secondInput: []sdk.DecCoins{},

			expectedOutput: []sdk.DecCoins{},
		},
		"[[100foo]] - [[50foo]]": {
			firstInput:  []sdk.DecCoins{sdk.NewDecCoins(hundredFoo)},
			secondInput: []sdk.DecCoins{sdk.NewDecCoins(fiftyFoo)},

			expectedOutput: []sdk.DecCoins{sdk.NewDecCoins(fiftyFoo)},
		},

		// error catching

		"different length inputs": {
			firstInput:  []sdk.DecCoins{hundredEach, hundredEach, hundredEach},
			secondInput: []sdk.DecCoins{fiftyEach, hundredEach},

			expectedOutput: []sdk.DecCoins{},
			expectError:    true,
		},

		"negative result": {
			firstInput:  []sdk.DecCoins{fiftyEach},
			secondInput: []sdk.DecCoins{hundredEach},

			expectedOutput:    []sdk.DecCoins{{sdk.DecCoin{Denom: "bar", Amount: osmomath.NewDec(-50)}, sdk.DecCoin{Denom: "foo", Amount: osmomath.NewDec(-50)}}},
			expectPanicUnsafe: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			var (
				actualOutput []sdk.DecCoins
				err1         error
			)
			osmoassert.ConditionalPanic(t, tc.expectPanicUnsafe, func() {
				actualOutput, err1 = osmoutils.SubDecCoinArrays(tc.firstInput, tc.secondInput)
			})

			actualOutputSafe, err2 := osmoutils.SafeSubDecCoinArrays(tc.firstInput, tc.secondInput)

			if tc.expectError {
				require.Error(t, err1)
				require.Error(t, err2)
				require.Equal(t, tc.expectedOutput, actualOutput)
				require.Equal(t, tc.expectedOutput, actualOutputSafe)
				return
			}

			require.NoError(t, err1)
			require.NoError(t, err2)
			if !tc.expectPanicUnsafe {
				require.Equal(t, tc.expectedOutput, actualOutput)
			}
			require.Equal(t, tc.expectedOutput, actualOutputSafe)
		})
	}
}

func TestAddDecCoins(t *testing.T) {
	tests := map[string]struct {
		firstInput  []sdk.DecCoins
		secondInput []sdk.DecCoins

		expectedOutput []sdk.DecCoins
		expectError    bool
	}{
		"[[100foo, 100bar], [100foo, 100bar]] + [[50foo, 50bar], [50foo, 100bar]]": {
			firstInput:  []sdk.DecCoins{hundredEach, hundredEach},
			secondInput: []sdk.DecCoins{fiftyEach, sdk.NewDecCoins(fiftyFoo, hundredBar)},

			expectedOutput: []sdk.DecCoins{
				hundredFiftyEach, // 100 + 50 on both coins
				{hundredBar.Add(hundredBar), hundredFiftyFoo}, // 100 + 50 on foo, 100 + 100 on bar (ordered lexicographically)
			},
		},
		// Flipped denom order
		"[[100bar, 100foo], [100foo, 100bar]] + [[50foo, 50bar], [50foo, 100bar]]": {
			firstInput: []sdk.DecCoins{
				sdk.NewDecCoins(hundredBar, hundredFoo),
				sdk.NewDecCoins(fiftyFoo, hundredBar)},
			secondInput: []sdk.DecCoins{fiftyEach, hundredEach},

			expectedOutput: []sdk.DecCoins{
				hundredFiftyEach, // 100 + 50 on both coins
				{hundredBar.Add(hundredBar), hundredFiftyFoo}, // 100 + 50 on foo, 100 + 100 on bar (ordered lexicographically)
			},
		},
		"both inputs empty": {
			firstInput:  []sdk.DecCoins{},
			secondInput: []sdk.DecCoins{},

			expectedOutput: []sdk.DecCoins{},
		},
		"[[100foo]] + [[50foo]]": {
			firstInput:  []sdk.DecCoins{sdk.NewDecCoins(hundredFoo)},
			secondInput: []sdk.DecCoins{sdk.NewDecCoins(fiftyFoo)},

			expectedOutput: []sdk.DecCoins{sdk.NewDecCoins(hundredFiftyFoo)},
		},

		// error catching

		"different length inputs": {
			firstInput:  []sdk.DecCoins{hundredEach, hundredEach, hundredEach},
			secondInput: []sdk.DecCoins{fiftyEach, hundredEach},

			expectedOutput: []sdk.DecCoins{},
			expectError:    true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualOutput, err := osmoutils.AddDecCoinArrays(tc.firstInput, tc.secondInput)

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

func TestCollapseDecCoinsArray(t *testing.T) {
	tests := map[string]struct {
		input []sdk.DecCoins

		expectedOutput sdk.DecCoins
	}{
		"[[100foo, 100bar], [100foo, 100bar]] -> [200foo, 200bar]": {
			input: []sdk.DecCoins{hundredEach, hundredEach},

			// Note that the order is lexicographic
			expectedOutput: sdk.NewDecCoins(hundredBar.Add(hundredBar), hundredFoo.Add(hundredFoo)),
		},
		// Note flipped denom order
		"[[100bar, 100foo], [50foo, 100bar]]": {
			input: []sdk.DecCoins{
				sdk.NewDecCoins(hundredBar, hundredFoo),
				sdk.NewDecCoins(fiftyFoo, hundredBar),
			},

			expectedOutput: sdk.NewDecCoins(twoHundredBar, hundredFiftyFoo),
		},
		"empty input array": {
			input: []sdk.DecCoins{},

			expectedOutput: sdk.DecCoins{},
		},
		"input array with nil DecCoins": {
			input: []sdk.DecCoins{emptyCoins, emptyCoins},

			expectedOutput: sdk.DecCoins(nil),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualOutput := osmoutils.CollapseDecCoinsArray(tc.input)
			require.Equal(t, tc.expectedOutput, actualOutput)
		})
	}
}

func TestConvertCoinsToDecCoins(t *testing.T) {
	tests := []struct {
		name             string
		inputCoins       sdk.Coins
		expectedDecCoins sdk.DecCoins
	}{
		{
			name:             "Empty input",
			inputCoins:       sdk.NewCoins(),
			expectedDecCoins: sdk.NewDecCoins(),
		},
		{
			name:             "Single coin",
			inputCoins:       sdk.NewCoins(sdk.NewCoin("atom", osmomath.NewInt(100000000))),
			expectedDecCoins: sdk.NewDecCoins(sdk.NewDecCoin("atom", osmomath.NewInt(100000000))),
		},
		{
			name:             "Multiple coins",
			inputCoins:       sdk.NewCoins(sdk.NewCoin("atom", osmomath.NewInt(100000000)), sdk.NewCoin("usdc", osmomath.NewInt(500000000))),
			expectedDecCoins: sdk.NewDecCoins(sdk.NewDecCoin("atom", osmomath.NewInt(100000000)), sdk.NewDecCoin("usdc", osmomath.NewInt(500000000))),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := osmoutils.ConvertCoinsToDecCoins(test.inputCoins)
			require.Equal(t, result, test.expectedDecCoins)

		})
	}
}
