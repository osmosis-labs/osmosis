package twap_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

var customGenesis = types.NewGenesisState(
	types.NewParams("week"),
	[]types.TwapRecord{
		{
			PoolId:                      1,
			Asset0Denom:                 "test1",
			Asset1Denom:                 "test2",
			Height:                      1,
			Time:                        baseTime,
			P0LastSpotPrice:             sdk.OneDec(),
			P1LastSpotPrice:             sdk.OneDec(),
			P0ArithmeticTwapAccumulator: sdk.OneDec(),
			P1ArithmeticTwapAccumulator: sdk.OneDec(),
		},
	})

// TestTWAPInitGenesis tests that genesis is initialized correctly
// with different parameters and state.
func (suite *TestSuite) TestTwapInitGenesis() {
	testCases := map[string]struct {
		twapGenesis *types.GenesisState

		expectPanic bool
	}{
		"default genesis - success": {
			twapGenesis: types.DefaultGenesis(),
		},
		"custom valid genesis - success": {
			twapGenesis: customGenesis,
		},
		"custom invalid genesis - error": {
			twapGenesis: types.NewGenesisState(
				types.NewParams("week"),
				[]types.TwapRecord{
					{
						PoolId:                      0, // invalid
						Asset0Denom:                 "test1",
						Asset1Denom:                 "test2",
						Height:                      1,
						Time:                        baseTime,
						P0LastSpotPrice:             sdk.OneDec(),
						P1LastSpotPrice:             sdk.OneDec(),
						P0ArithmeticTwapAccumulator: sdk.OneDec(),
						P1ArithmeticTwapAccumulator: sdk.OneDec(),
					},
				}),

			expectPanic: true,
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.Setup()
			// Setup.
			ctx := suite.Ctx
			twapKeeper := suite.App.TwapKeeper

			// Test.
			osmoassert.ConditionalPanic(suite.T(), tc.expectPanic, func() { twapKeeper.InitGenesis(ctx, tc.twapGenesis) })
			if tc.expectPanic {
				return
			}

			// Assertions.

			// Parameters were set.
			suite.Require().Equal(tc.twapGenesis.Params, twapKeeper.GetParams(ctx))
		})
	}
}

// TestTWAPExportGenesis tests that genesis is exported correctly.
// It first initializes genesis to the expected value. Then, attempts
// to export it. Lastly, compares exported to the expected.
func (suite *TestSuite) TestTWAPExportGenesis() {
	testCases := map[string]struct {
		expectedGenesis *types.GenesisState
	}{
		"default genesis": {
			expectedGenesis: types.DefaultGenesis(),
		},
		"custom genesis": {
			expectedGenesis: customGenesis,
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			// Setup.
			app := suite.App
			ctx := suite.Ctx
			twapKeeper := app.TwapKeeper

			twapKeeper.InitGenesis(ctx, tc.expectedGenesis)

			// Test.
			actualGenesis := twapKeeper.ExportGenesis(ctx)

			// Assertions.
			suite.Require().Equal(tc.expectedGenesis, actualGenesis)
		})
	}
}
