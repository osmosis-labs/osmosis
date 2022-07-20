package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/osmosis-labs/osmosis/v10/osmoutils"
	"github.com/osmosis-labs/osmosis/v10/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v10/x/mint/types"
)

var customGenesis = types.NewGenesisState(
	types.NewMinter(sdk.ZeroDec()),
	types.NewParams(
		"uosmo",                  // denom
		sdk.NewDec(200),          // epoch provisions
		"year",                   // epoch identifier
		sdk.NewDecWithPrec(5, 1), // Halven factor
		5,                        // Halven perion in epochs
		types.DistributionProportions{
			Staking:          sdk.NewDecWithPrec(25, 2),
			PoolIncentives:   sdk.NewDecWithPrec(25, 2),
			DeveloperRewards: sdk.NewDecWithPrec(25, 2),
			CommunityPool:    sdk.NewDecWithPrec(25, 2),
		},
		[]types.WeightedAddress{
			{
				Address: "osmo14kjcwdwcqsujkdt8n5qwpd8x8ty2rys5rjrdjj",
				Weight:  sdk.NewDecWithPrec(6, 1),
			},
			{
				Address: "osmo1gw445ta0aqn26suz2rg3tkqfpxnq2hs224d7gq",
				Weight:  sdk.NewDecWithPrec(4, 1),
			},
		},
		2), // minting reward distribution start epoch
	3) // halven started epoch

// TestMintInitGenesis tests that genesis is initialized correctly
// with different parameters and state.
func (suite *KeeperTestSuite) TestMintInitGenesis() {
	testCases := map[string]struct {
		mintGenesis                     *types.GenesisState
		mintDenom                       string
		ctxHeight                       int64
		isDeveloperModuleAccountCreated bool

		expectPanic             bool
		expectedEpochProvisions sdk.Dec
		// Deltas represent by how much a certain paramets
		// has changed after calling InitGenesis()
		expectedSupplyOffsetDelta           sdk.Int
		expectedSupplyWithOffsetDelta       sdk.Int
		expectedDeveloperVestingAmountDelta sdk.Int
		expectedHalvenStartedEpoch          int64
	}{
		"default genesis - developer module account is not created prior to InitGenesis() - created during the call": {
			mintGenesis: types.DefaultGenesisState(),
			mintDenom:   sdk.DefaultBondDenom,

			expectedEpochProvisions:             types.DefaultGenesisState().Params.GenesisEpochProvisions,
			expectedSupplyOffsetDelta:           sdk.NewInt(keeper.DeveloperVestingAmount).Neg(),
			expectedSupplyWithOffsetDelta:       sdk.ZeroInt(),
			expectedDeveloperVestingAmountDelta: sdk.NewInt(keeper.DeveloperVestingAmount),
		},
		"default genesis - developer module account is created prior to InitGenesis() - not created during the call": {
			mintGenesis:                     types.DefaultGenesisState(),
			mintDenom:                       sdk.DefaultBondDenom,
			isDeveloperModuleAccountCreated: true,

			expectedEpochProvisions:             types.DefaultGenesisState().Params.GenesisEpochProvisions,
			expectedSupplyOffsetDelta:           sdk.ZeroInt(),
			expectedSupplyWithOffsetDelta:       sdk.ZeroInt(),
			expectedDeveloperVestingAmountDelta: sdk.ZeroInt(),
		},
		"custom genesis": {
			mintGenesis: customGenesis,
			mintDenom:   "uosmo",

			expectedEpochProvisions:             sdk.NewDec(200),
			expectedSupplyOffsetDelta:           sdk.NewInt(keeper.DeveloperVestingAmount).Neg(),
			expectedSupplyWithOffsetDelta:       sdk.ZeroInt(),
			expectedDeveloperVestingAmountDelta: sdk.NewInt(keeper.DeveloperVestingAmount),
			expectedHalvenStartedEpoch:          3,
		},
		"nil genesis state - panic": {
			mintDenom:   sdk.DefaultBondDenom,
			expectPanic: true,
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			// Setup.
			suite.setupDeveloperVestingModuleAccountTest(tc.ctxHeight, tc.isDeveloperModuleAccountCreated)
			ctx := suite.Ctx
			accountKeeper := suite.App.AccountKeeper
			bankKeeper := suite.App.BankKeeper
			mintKeeper := suite.App.MintKeeper

			developerAccount := accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)

			originalSupplyOffset := bankKeeper.GetSupplyOffset(ctx, tc.mintDenom)
			originalSupplyWithOffset := bankKeeper.GetSupplyWithOffset(ctx, tc.mintDenom)
			originalVestingCoins := bankKeeper.GetBalance(ctx, developerAccount, tc.mintDenom)

			// Test.
			osmoutils.ConditionalPanic(suite.T(), tc.expectPanic, func() { mintKeeper.InitGenesis(ctx, tc.mintGenesis) })
			if tc.expectPanic {
				return
			}

			// Assertions.

			// Module account was created.
			acc := accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
			suite.NotNil(acc)

			// Epoch provisions are set to genesis epoch provisions from params.
			actualEpochProvisions := mintKeeper.GetMinter(ctx).EpochProvisions
			suite.Require().Equal(tc.expectedEpochProvisions, actualEpochProvisions)

			// Supply offset is applied to genesis supply.
			actualSupplyOffset := bankKeeper.GetSupplyOffset(ctx, tc.mintDenom)
			expectedSupplyOffset := tc.expectedSupplyOffsetDelta.Add(originalSupplyOffset)
			suite.Require().Equal(expectedSupplyOffset, actualSupplyOffset)

			// Supply with offset is as expected.
			actualSupplyWithOffset := bankKeeper.GetSupplyWithOffset(ctx, tc.mintDenom).Amount
			expectedSupplyWithOffset := tc.expectedSupplyWithOffsetDelta.Add(originalSupplyWithOffset.Amount)
			suite.Require().Equal(expectedSupplyWithOffset.Int64(), actualSupplyWithOffset.Int64())

			// Developer vesting account has the desired amount of tokens.
			actualVestingCoins := bankKeeper.GetBalance(ctx, developerAccount, tc.mintDenom)
			expectedDeveloperVestingAmount := tc.expectedDeveloperVestingAmountDelta.Add(originalVestingCoins.Amount)
			suite.Require().Equal(expectedDeveloperVestingAmount.Int64(), actualVestingCoins.Amount.Int64())

			// Last halven epoch num is set to 0.
			suite.Require().Equal(tc.expectedHalvenStartedEpoch, mintKeeper.GetLastHalvenEpochNum(ctx))
		})
	}
}

// TestMintExportGenesis tests that genesis is exported correctly.
// It first initializes genesis to the expected value. Then, attempts
// to export it. Lastly, compares exported to the expected.
func (suite *KeeperTestSuite) TestMintExportGenesis() {
	testCases := map[string]struct {
		expectedGenesis *types.GenesisState
	}{
		"default genesis": {
			expectedGenesis: types.DefaultGenesisState(),
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

			app.MintKeeper.InitGenesis(ctx, tc.expectedGenesis)

			// Test.
			actualGenesis := app.MintKeeper.ExportGenesis(ctx)

			// Assertions.
			suite.Require().Equal(tc.expectedGenesis, actualGenesis)
		})
	}
}
