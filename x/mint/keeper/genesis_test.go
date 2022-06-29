package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/osmosis-labs/osmosis/v7/app"

	"github.com/osmosis-labs/osmosis/v7/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/mint/types"
)

func TestMintGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// TestMintInitGenesis test that genesis is initialized correctly.
func (suite *KeeperTestSuite) TestMintInitGenesis() {
	testcases := map[string]struct {
		mintGenesis                     *types.GenesisState
		mintDenom                       string
		ctxHeight                       int64
		isDeveloperModuleAccountCreated bool

		expectPanic             bool
		expectedEpochProvisions sdk.Dec
		// Deltas represent by how much a certain paramets
		// has changeda after calling InitGenesis()
		expectedSupplyOffsetDelta           sdk.Int
		expectedSupplyWithOffsetDelta       sdk.Int
		expectedDeveloperVestingAmountDelta sdk.Int
		expectedHalvenStartedEpoch          int64
	}{
		"default genesis - developer module account is not created priot to InitGenesis() - created during the call": {
			mintGenesis: types.DefaultGenesisState(),
			mintDenom:   sdk.DefaultBondDenom,

			expectedEpochProvisions:             types.DefaultGenesisState().Params.GenesisEpochProvisions,
			expectedSupplyOffsetDelta:           sdk.NewInt(keeper.DeveloperVestingAmount).Neg(),
			expectedSupplyWithOffsetDelta:       sdk.ZeroInt(),
			expectedDeveloperVestingAmountDelta: sdk.NewInt(keeper.DeveloperVestingAmount),
		},
		"default genesis - developer module account is created priot to InitGenesis() - not created during the call": {
			mintGenesis:                     types.DefaultGenesisState(),
			mintDenom:                       sdk.DefaultBondDenom,
			isDeveloperModuleAccountCreated: true,

			expectedEpochProvisions:             types.DefaultGenesisState().Params.GenesisEpochProvisions,
			expectedSupplyOffsetDelta:           sdk.ZeroInt(),
			expectedSupplyWithOffsetDelta:       sdk.ZeroInt(),
			expectedDeveloperVestingAmountDelta: sdk.ZeroInt(),
		},
		"custom genesis": {
			mintGenesis: types.NewGenesisState(
				types.NewMinter(sdk.ZeroDec()),
				types.NewParams(
					"uosmo",         // denom
					sdk.NewDec(200), // epoch provisions
					"year",
					sdk.NewDecWithPrec(5, 1),
					5,
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
					2),
				3), // halven started epoch
			mintDenom: "uosmo",

			expectedEpochProvisions:             sdk.NewDec(200),
			expectedSupplyOffsetDelta:           sdk.NewInt(keeper.DeveloperVestingAmount).Neg(),
			expectedSupplyWithOffsetDelta:       sdk.ZeroInt(),
			expectedDeveloperVestingAmountDelta: sdk.NewInt(keeper.DeveloperVestingAmount),
			expectedHalvenStartedEpoch:          3,
		},
		"non-zero ctx height - panic": {
			mintGenesis: types.DefaultGenesisState(),
			mintDenom:   sdk.DefaultBondDenom,
			ctxHeight:   1,

			expectPanic: true,
		},
		"nil genesis state - panic": {
			mintDenom:   sdk.DefaultBondDenom,
			expectPanic: true,
		},
	}

	// Sets up each test case by reverting some default logic added by suite.Setup()
	// Specifically, it removes the module account
	// from account keeper if isDeveloperModuleAccountCreated is true.
	// Returns initialized context to be used in tests.
	testcaseSetup := func(suite *KeeperTestSuite, blockHeight int64, isDeveloperModuleAccountCreated bool) sdk.Context {
		suite.Setup()
		// Reset height to the desired value since test suite setup initialized
		// it to 1.
		ctx := suite.Ctx.WithBlockHeader(tmproto.Header{Height: blockHeight})

		// If developer module account is created, the suite.Setup() also correctly sets the offset
		if !isDeveloperModuleAccountCreated {
			// Remove the developer vesting account since suite setup creates and initializes it.
			// This environment w/o the developer vesting account configured is necessary for
			// testing edge cases of multiple tests.
			developerVestingAccount := suite.App.AccountKeeper.GetAccount(ctx, suite.App.AccountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName))
			suite.App.AccountKeeper.RemoveAccount(ctx, developerVestingAccount)
			suite.App.BankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(keeper.DeveloperVestingAmount))))

			// Reset supply offset to 0
			supplyOffset := suite.App.BankKeeper.GetSupplyOffset(ctx, sdk.DefaultBondDenom)
			suite.App.BankKeeper.AddSupplyOffset(ctx, sdk.DefaultBondDenom, supplyOffset.Mul(sdk.NewInt(-1)))
			suite.Equal(sdk.ZeroInt(), suite.App.BankKeeper.GetSupplyOffset(ctx, sdk.DefaultBondDenom))
		}

		return ctx
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			ctx := testcaseSetup(suite, tc.ctxHeight, tc.isDeveloperModuleAccountCreated)

			developerAccount := suite.App.AccountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)

			originalSupplyOffset := suite.App.BankKeeper.GetSupplyOffset(ctx, tc.mintDenom)
			originalSupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(ctx, tc.mintDenom)
			originalVestingCoins := suite.App.BankKeeper.GetBalance(ctx, developerAccount, tc.mintDenom)

			if tc.expectPanic {
				suite.Panics(func() {
					suite.App.MintKeeper.InitGenesis(ctx, suite.App.AccountKeeper, suite.App.BankKeeper, tc.mintGenesis)
				})
				return
			}

			suite.NotPanics(func() {
				suite.App.MintKeeper.InitGenesis(ctx, suite.App.AccountKeeper, suite.App.BankKeeper, tc.mintGenesis)
			})

			// Epoch provisions are set to genesis epoch provisions from params.
			actualEpochProvisions := suite.App.MintKeeper.GetMinter(ctx).EpochProvisions
			suite.Equal(tc.expectedEpochProvisions, actualEpochProvisions)

			// Supply offset is applied to genesis supply.
			actualSupplyOffset := suite.App.BankKeeper.GetSupplyOffset(ctx, tc.mintDenom)
			expectedSupplyOffset := tc.expectedSupplyOffsetDelta.Add(originalSupplyOffset)
			suite.Equal(expectedSupplyOffset, actualSupplyOffset)

			// Supply with offset is as expected.
			actualSupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(ctx, tc.mintDenom).Amount
			expectedSupplyWithOffset := tc.expectedSupplyWithOffsetDelta.Add(originalSupplyWithOffset.Amount)
			suite.Equal(expectedSupplyWithOffset.Int64(), actualSupplyWithOffset.Int64())

			// Developer vesting account has the desired amount of tokens.
			actualVestingCoins := suite.App.BankKeeper.GetBalance(ctx, developerAccount, tc.mintDenom)
			expectedDeveloperVestingAmount := tc.expectedDeveloperVestingAmountDelta.Add(originalVestingCoins.Amount)
			suite.Equal(expectedDeveloperVestingAmount.Int64(), actualVestingCoins.Amount.Int64())

			// Last halven epoch num is set to 0.
			suite.Equal(tc.expectedHalvenStartedEpoch, suite.App.MintKeeper.GetLastHalvenEpochNum(ctx))
		})
	}
}

// TestMintInitGenesis test that genesis is initialized correctly.
func (suite *KeeperTestSuite) TestMintInitGenesis_ModuleAccountCreated() {
	const developerVestingAmount = 225000000000000

	// InitGenesis occurs in app setup.
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// Epoch provisions are set to genesis epoch provisions from params.
	epochProvisions := app.MintKeeper.GetMinter(ctx).EpochProvisions
	suite.Equal(epochProvisions, types.DefaultParams().GenesisEpochProvisions)

	// Supply offset is applied to genesis supply.
	expectedSupplyWithOffset := int64(0)
	actualSupplyWithOffset := app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount.Int64()
	suite.Equal(expectedSupplyWithOffset, actualSupplyWithOffset)

	// Developer vesting account has the desired amount of tokens.
	expectedVestingCoins := sdk.NewInt(developerVestingAmount)
	developerAccount := app.AccountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	initialVestingCoins := app.BankKeeper.GetBalance(ctx, developerAccount, sdk.DefaultBondDenom)
	suite.Equal(expectedVestingCoins, initialVestingCoins.Amount)

	// Last halven epoch num is set to 0.
	suite.Equal(int64(0), app.MintKeeper.GetLastHalvenEpochNum(ctx))
}

// TestMintExportGenesis test that genesis is exported correctly.
func (suite *KeeperTestSuite) TestMintInitAndExportGenesis_InverseRelationship() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	const expectedLastHalvenEpochNum = 1

	var expectedEpochProvisions = sdk.NewDec(2)

	// change last halven epoch num to non-zero.
	app.MintKeeper.SetLastHalvenEpochNum(ctx, expectedLastHalvenEpochNum)

	// Change epoch provisions to non-default params value.
	app.MintKeeper.SetMinter(ctx, types.NewMinter(expectedEpochProvisions))

	// Modify changed values on the exported genesis.
	expectedGenesis := types.DefaultGenesisState()
	expectedGenesis.HalvenStartedEpoch = expectedLastHalvenEpochNum
	expectedGenesis.Minter.EpochProvisions = expectedEpochProvisions

	actualGenesis := app.MintKeeper.ExportGenesis(ctx)

	suite.Equal(expectedGenesis, actualGenesis)
}
