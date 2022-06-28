package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/osmosis-labs/osmosis/v7/app"

	"github.com/osmosis-labs/osmosis/v7/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/mint/types"
)

// TestMintInitGenesis test that genesis is initialized correctly.
func TestMintInitGenesis(t *testing.T) {
	testcases := map[string]struct {
		mintGenesis *types.GenesisState

		isDeveloperModuleAccountCreated bool
		expectedDeveloperVersingAmount  sdk.Int
	}{
		"default genesis - developer module account not created": {
			mintGenesis: types.DefaultGenesisState(),

			expectedDeveloperVersingAmount: sdk.NewInt(keeper.DeveloperVestingAmount),
		},
		"defaul genesis - developer module account created": {
			expectedDeveloperVersingAmount:  sdk.NewInt(5),
			isDeveloperModuleAccountCreated: true,
		},
		"custom genesis": {
			mintGenesis:                    types.DefaultGenesisState(),
			expectedDeveloperVersingAmount: sdk.NewInt(keeper.DeveloperVestingAmount),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {

			// InitGenesis occurs in app setup.
			app := simapp.Setup(false)
			ctx := app.BaseApp.NewContext(false, tmproto.Header{})

			if tc.isDeveloperModuleAccountCreated {
				require.NoError(t, app.MintKeeper.CreateDeveloperVestingModuleAccount(ctx, sdk.NewCoin(sdk.DefaultBondDenom, tc.expectedDeveloperVersingAmount)))
				app.MintKeeper.SetInitialSupplyOffsetDuringMigration(ctx)
			}

			app.MintKeeper.InitGenesis(ctx, app.AccountKeeper, app.BankKeeper, types.DefaultGenesisState())

			// Epoch provisions are set to genesis epoch provisions from params.
			epochProvisions := app.MintKeeper.GetMinter(ctx).EpochProvisions
			require.Equal(t, epochProvisions, types.DefaultParams().GenesisEpochProvisions)

			// Supply offset is applied to genesis supply.
			expectedSupplyWithOffset := int64(0)
			actualSupplyWithOffset := app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount.Int64()
			require.Equal(t, expectedSupplyWithOffset, actualSupplyWithOffset)

			// Developer vesting account has the desired amount of tokens.
			developerAccount := app.AccountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
			initialVestingCoins := app.BankKeeper.GetBalance(ctx, developerAccount, sdk.DefaultBondDenom)
			require.Equal(t, tc.expectedDeveloperVersingAmount, initialVestingCoins.Amount)

			// Last halven epoch num is set to 0.
			require.Equal(t, int64(0), app.MintKeeper.GetLastHalvenEpochNum(ctx))
		})
	}
}

// TestMintInitGenesis test that genesis is initialized correctly.
func (suite *KeeperTestSuite) TestMintInitGenesis_ModuleAccountCreated(t *testing.T) {
	const developerVestingAmount = 225000000000000

	// InitGenesis occurs in app setup.
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// Epoch provisions are set to genesis epoch provisions from params.
	epochProvisions := app.MintKeeper.GetMinter(ctx).EpochProvisions
	require.Equal(t, epochProvisions, types.DefaultParams().GenesisEpochProvisions)

	// Supply offset is applied to genesis supply.
	expectedSupplyWithOffset := int64(0)
	actualSupplyWithOffset := app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount.Int64()
	require.Equal(t, expectedSupplyWithOffset, actualSupplyWithOffset)

	// Developer vesting account has the desired amount of tokens.
	expectedVestingCoins := sdk.NewInt(developerVestingAmount)
	developerAccount := app.AccountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	initialVestingCoins := app.BankKeeper.GetBalance(ctx, developerAccount, sdk.DefaultBondDenom)
	require.Equal(t, expectedVestingCoins, initialVestingCoins.Amount)

	// Last halven epoch num is set to 0.
	require.Equal(t, int64(0), app.MintKeeper.GetLastHalvenEpochNum(ctx))
}

// TestMintExportGenesis test that genesis is exported correctly.
func TestMintInitAndExportGenesis(t *testing.T) {
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

	require.Equal(t, expectedGenesis, actualGenesis)
}
