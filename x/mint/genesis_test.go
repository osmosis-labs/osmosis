package mint_test

import (
	"testing"

	simapp "github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/x/mint/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMintInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	validateGenesis := types.ValidateGenesis(*types.DefaultGenesisState())
	require.NoError(t, validateGenesis)

	developerAccount := app.AccountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	initialVestingCoins := app.BankKeeper.GetBalance(ctx, developerAccount, sdk.DefaultBondDenom)

	expectedVestingCoins, ok := sdk.NewIntFromString("225000000000000")
	require.True(t, ok)
	require.Equal(t, expectedVestingCoins, initialVestingCoins.Amount)
	require.Equal(t, int64(0), app.MintKeeper.GetLastHalvenEpochNum(ctx))
}
