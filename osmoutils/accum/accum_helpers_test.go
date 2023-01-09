package accum_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
)

func (suite *AccumTestSuite) TestValidateAccumulatorValue() {
	tests := map[string]struct {
		customAccumulatorValue      sdk.DecCoins
		oldPositionAccumulatorValue sdk.DecCoins
		expectError                 error
	}{
		"negative custom coins - error": {
			customAccumulatorValue:      initialCoinsDenomOne.MulDec(sdk.NewDec(-1)),
			oldPositionAccumulatorValue: emptyCoins,
			expectError:                 accum.NegativeCustomAccError{initialCoinsDenomOne.MulDec(sdk.NewDec(-1))},
		},
		"old accumulator coins are greater than new - error": {
			customAccumulatorValue:      initialCoinsDenomOne,
			oldPositionAccumulatorValue: initialCoinsDenomOne.Add(sdk.NewDecCoin(initialCoinDenomOne.Denom, sdk.OneInt())),
			expectError:                 accum.NegativeAccDifferenceError{sdk.NewDecCoins(sdk.NewDecCoin(initialCoinDenomOne.Denom, sdk.OneInt()))},
		},
		"old accumulator coins are a superset of new - error": {
			customAccumulatorValue:      initialCoinsDenomOne,
			oldPositionAccumulatorValue: initialCoinsDenomOne.Add(initialCoinDenomTwo),
			expectError:                 accum.NegativeAccDifferenceError{sdk.NewDecCoins(initialCoinDenomTwo)},
		},
		"new accumulator coins are a superset of old - success": {
			customAccumulatorValue:      initialCoinsDenomOne.Add(initialCoinDenomTwo),
			oldPositionAccumulatorValue: initialCoinsDenomOne,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			suite.SetupTest()

			err := accum.ValidateAccumulatorValue(tc.customAccumulatorValue, tc.oldPositionAccumulatorValue)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectError, err)
				return
			}
			suite.Require().NoError(err)
		})
	}
}
