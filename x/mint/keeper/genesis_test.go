package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/mint/types"
)

var customGenesis = types.NewGenesisState(
	types.NewMinter(osmomath.ZeroDec()),
	types.NewParams(
		appparams.BaseCoinUnit,        // denom
		osmomath.NewDec(200),          // epoch provisions
		"year",                        // epoch identifier
		osmomath.NewDecWithPrec(5, 1), // reduction factor
		5,                             // reduction perion in epochs
		types.DistributionProportions{
			Staking:          osmomath.NewDecWithPrec(25, 2),
			PoolIncentives:   osmomath.NewDecWithPrec(25, 2),
			DeveloperRewards: osmomath.NewDecWithPrec(25, 2),
			CommunityPool:    osmomath.NewDecWithPrec(25, 2),
		},
		[]types.WeightedAddress{
			{
				Address: "osmo14kjcwdwcqsujkdt8n5qwpd8x8ty2rys5rjrdjj",
				Weight:  osmomath.NewDecWithPrec(6, 1),
			},
			{
				Address: "osmo1gw445ta0aqn26suz2rg3tkqfpxnq2hs224d7gq",
				Weight:  osmomath.NewDecWithPrec(4, 1),
			},
		},
		2), // minting reward distribution start epoch
	3) // halven started epoch

// TestMintInitGenesis tests that genesis is initialized correctly
// with different parameters and state.
func (s *KeeperTestSuite) TestMintInitGenesis() {
	testCases := map[string]struct {
		mintGenesis                     *types.GenesisState
		mintDenom                       string
		ctxHeight                       int64
		isDeveloperModuleAccountCreated bool

		expectPanic             bool
		expectedEpochProvisions osmomath.Dec
		// Deltas represent by how much a certain paramets
		// has changed after calling InitGenesis()
		expectedSupplyOffsetDelta           osmomath.Int
		expectedSupplyWithOffsetDelta       osmomath.Int
		expectedDeveloperVestingAmountDelta osmomath.Int
		expectedHalvenStartedEpoch          int64
	}{
		"default genesis - developer module account is not created prior to InitGenesis() - created during the call": {
			mintGenesis: types.DefaultGenesisState(),
			mintDenom:   sdk.DefaultBondDenom,

			expectedEpochProvisions:             types.DefaultGenesisState().Params.GenesisEpochProvisions,
			expectedSupplyOffsetDelta:           osmomath.NewInt(keeper.DeveloperVestingAmount).Neg(),
			expectedSupplyWithOffsetDelta:       osmomath.ZeroInt(),
			expectedDeveloperVestingAmountDelta: osmomath.NewInt(keeper.DeveloperVestingAmount),
		},
		"default genesis - developer module account is created prior to InitGenesis() - not created during the call": {
			mintGenesis:                     types.DefaultGenesisState(),
			mintDenom:                       sdk.DefaultBondDenom,
			isDeveloperModuleAccountCreated: true,

			expectedEpochProvisions:             types.DefaultGenesisState().Params.GenesisEpochProvisions,
			expectedSupplyOffsetDelta:           osmomath.ZeroInt(),
			expectedSupplyWithOffsetDelta:       osmomath.ZeroInt(),
			expectedDeveloperVestingAmountDelta: osmomath.ZeroInt(),
		},
		"custom genesis": {
			mintGenesis: customGenesis,
			mintDenom:   appparams.BaseCoinUnit,

			expectedEpochProvisions:             osmomath.NewDec(200),
			expectedSupplyOffsetDelta:           osmomath.NewInt(keeper.DeveloperVestingAmount).Neg(),
			expectedSupplyWithOffsetDelta:       osmomath.ZeroInt(),
			expectedDeveloperVestingAmountDelta: osmomath.NewInt(keeper.DeveloperVestingAmount),
			expectedHalvenStartedEpoch:          3,
		},
		"nil genesis state - panic": {
			mintDenom:   sdk.DefaultBondDenom,
			expectPanic: true,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			// Setup.
			s.setupDeveloperVestingModuleAccountTest(tc.ctxHeight, tc.isDeveloperModuleAccountCreated)
			ctx := s.Ctx
			accountKeeper := s.App.AccountKeeper
			bankKeeper := s.App.BankKeeper
			mintKeeper := s.App.MintKeeper

			developerAccount := accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)

			originalSupplyOffset := bankKeeper.GetSupplyOffset(ctx, tc.mintDenom)
			originalSupplyWithOffset := bankKeeper.GetSupplyWithOffset(ctx, tc.mintDenom)
			originalVestingCoins := bankKeeper.GetBalance(ctx, developerAccount, tc.mintDenom)

			// Test.
			osmoassert.ConditionalPanic(s.T(), tc.expectPanic, func() { mintKeeper.InitGenesis(ctx, tc.mintGenesis) })
			if tc.expectPanic {
				return
			}

			// Assertions.

			// Module account was created.
			acc := accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
			s.NotNil(acc)

			// Epoch provisions are set to genesis epoch provisions from params.
			actualEpochProvisions := mintKeeper.GetMinter(ctx).EpochProvisions
			s.Require().Equal(tc.expectedEpochProvisions, actualEpochProvisions)

			// Supply offset is applied to genesis supply.
			actualSupplyOffset := bankKeeper.GetSupplyOffset(ctx, tc.mintDenom)
			expectedSupplyOffset := tc.expectedSupplyOffsetDelta.Add(originalSupplyOffset)
			s.Require().Equal(expectedSupplyOffset, actualSupplyOffset)

			// Supply with offset is as expected.
			actualSupplyWithOffset := bankKeeper.GetSupplyWithOffset(ctx, tc.mintDenom).Amount
			expectedSupplyWithOffset := tc.expectedSupplyWithOffsetDelta.Add(originalSupplyWithOffset.Amount)
			s.Require().Equal(expectedSupplyWithOffset.Int64(), actualSupplyWithOffset.Int64())

			// Developer vesting account has the desired amount of tokens.
			actualVestingCoins := bankKeeper.GetBalance(ctx, developerAccount, tc.mintDenom)
			expectedDeveloperVestingAmount := tc.expectedDeveloperVestingAmountDelta.Add(originalVestingCoins.Amount)
			s.Require().Equal(expectedDeveloperVestingAmount.Int64(), actualVestingCoins.Amount.Int64())

			// Last halven epoch num is set to 0.
			s.Require().Equal(tc.expectedHalvenStartedEpoch, mintKeeper.GetLastReductionEpochNum(ctx))
		})
	}
}

// TestMintExportGenesis tests that genesis is exported correctly.
// It first initializes genesis to the expected value. Then, attempts
// to export it. Lastly, compares exported to the expected.
func (s *KeeperTestSuite) TestMintExportGenesis() {
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
		s.Run(name, func() {
			// Setup.
			app := s.App
			ctx := s.Ctx

			app.MintKeeper.InitGenesis(ctx, tc.expectedGenesis)

			// Test.
			actualGenesis := app.MintKeeper.ExportGenesis(ctx)

			// Assertions.
			s.Require().Equal(tc.expectedGenesis, actualGenesis)
		})
	}
}
