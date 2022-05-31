package keeper_test

import (
	"testing"

<<<<<<< HEAD:x/mint/genesis_test.go
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/osmosis-labs/osmosis/v10/app"
	"github.com/osmosis-labs/osmosis/v10/x/mint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
=======
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/osmosis-labs/osmosis/v7/app"

	"github.com/osmosis-labs/osmosis/v7/x/mint/types"
>>>>>>> 61a207f8 (chore: move init export genesis to keepers (#1631)):x/mint/keeper/genesis_test.go
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
