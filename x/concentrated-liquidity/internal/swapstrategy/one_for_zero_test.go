package swapstrategy_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/internal/swapstrategy"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

func (suite *StrategyTestSuite) TestGetSqrtTargetPrice() {
	var (
		two   = sdk.NewDec(2)
		three = sdk.NewDec(2)
		four  = sdk.NewDec(4)
		five  = sdk.NewDec(5)
	)

	tests := map[string]struct {
		isZeroForOne      bool
		sqrtPriceLimit    sdk.Dec
		nextTickSqrtPrice sdk.Dec
		expectedResult    sdk.Dec
	}{
		"zero for one: nextTickSqrtPrice == sqrtPriceLimit -> no-op": {
			isZeroForOne:      true,
			sqrtPriceLimit:    sdk.OneDec(),
			nextTickSqrtPrice: sdk.OneDec(),
			expectedResult:    sdk.OneDec(),
		},
		"zero for one: nextTickSqrtPrice > sqrtPriceLimit -> nextTickSqrtPrice": {
			isZeroForOne:      true,
			sqrtPriceLimit:    three,
			nextTickSqrtPrice: four,
			expectedResult:    four,
		},
		"zero for one: nextTickSqrtPrice < sqrtPriceLimit -> sqrtPriceLimit": {
			isZeroForOne:      true,
			sqrtPriceLimit:    five,
			nextTickSqrtPrice: two,
			expectedResult:    five,
		},
		"one for zero: nextTickSqrtPrice == sqrtPriceLimit -> no-op": {
			isZeroForOne:      false,
			sqrtPriceLimit:    sdk.OneDec(),
			nextTickSqrtPrice: sdk.OneDec(),
			expectedResult:    sdk.OneDec(),
		},
		"one for zero: nextTickSqrtPrice > sqrtPriceLimit -> sqrtPriceLimit": {
			isZeroForOne:      false,
			sqrtPriceLimit:    three,
			nextTickSqrtPrice: four,
			expectedResult:    three,
		},
		"zero for one: nextTickSqrtPrice < sqrtPriceLimit -> nextTickSqrtPrice": {
			isZeroForOne:      false,
			sqrtPriceLimit:    five,
			nextTickSqrtPrice: two,
			expectedResult:    two,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			sut := swapstrategy.New(tc.isZeroForOne, false, tc.sqrtPriceLimit, suite.App.GetKey(types.ModuleName), sdk.ZeroDec())

			actualSqrtTargetPrice := sut.GetSqrtTargetPrice(tc.nextTickSqrtPrice)

			suite.Require().Equal(tc.expectedResult, actualSqrtTargetPrice)

		})
	}
}
