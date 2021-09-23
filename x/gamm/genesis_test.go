package gamm_test

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/osmosis-labs/osmosis/app"
	appparams "github.com/osmosis-labs/osmosis/app/params"
	"github.com/osmosis-labs/osmosis/x/gamm"
	"github.com/osmosis-labs/osmosis/x/gamm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestGammInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	poolI, err := types.NewPool(1, types.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.NewDecWithPrec(1, 2),
	}, []types.PoolAsset{
		{
			Weight: sdk.NewInt(1),
			Token:  sdk.NewInt64Coin(sdk.DefaultBondDenom, 10),
		},
		{
			Weight: sdk.NewInt(1),
			Token:  sdk.NewInt64Coin("nodetoken", 10),
		},
	}, "", ctx.BlockTime())
	require.NoError(t, err)

	pool, ok := poolI.(*types.Pool)
	require.True(t, ok)

	any, err := codectypes.NewAnyWithValue(pool)
	require.NoError(t, err)

	gamm.InitGenesis(ctx, app.GAMMKeeper, types.GenesisState{
		Pools:          []*codectypes.Any{any},
		NextPoolNumber: 2,
		Params: types.Params{
			PoolCreationFee: sdk.Coins{sdk.NewInt64Coin(appparams.BaseCoinUnit, 1000_000_000)},
		},
	}, app.AppCodec())

	require.Equal(t, app.GAMMKeeper.GetNextPoolNumber(ctx), uint64(2))
	poolStored, err := app.GAMMKeeper.GetPool(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, pool.GetId(), poolStored.GetId())
	require.Equal(t, pool.GetAddress(), poolStored.GetAddress())
	require.Equal(t, pool.GetPoolParams(), poolStored.GetPoolParams())
	require.Equal(t, pool.GetTotalWeight(), poolStored.GetTotalWeight())
	require.Equal(t, pool.GetTotalShares(), poolStored.GetTotalShares())
	require.Equal(t, pool.GetAllPoolAssets(), poolStored.GetAllPoolAssets())
	require.Equal(t, pool.String(), poolStored.String())

	_, err = app.GAMMKeeper.GetPool(ctx, 2)
	require.Error(t, err)

	liquidity := app.GAMMKeeper.GetTotalLiquidity(ctx)
	require.Equal(t, liquidity, sdk.Coins{sdk.NewInt64Coin("nodetoken", 10), sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)})
}

func TestGammExportGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	app.BankKeeper.SetBalances(ctx, acc1, sdk.Coins{
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewInt64Coin("foo", 100000),
		sdk.NewInt64Coin("bar", 100000),
	})

	_, err := app.GAMMKeeper.CreatePool(ctx, acc1, types.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.NewDecWithPrec(1, 2),
	}, []types.PoolAsset{{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}, {
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}}, "")
	require.NoError(t, err)

	_, err = app.GAMMKeeper.CreatePool(ctx, acc1, types.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.NewDecWithPrec(1, 2),
	}, []types.PoolAsset{{
		Weight: sdk.NewInt(70),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}, {
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}}, "")
	require.NoError(t, err)

	genesis := gamm.ExportGenesis(ctx, app.GAMMKeeper)
	require.Equal(t, genesis.NextPoolNumber, uint64(3))
	require.Len(t, genesis.Pools, 2)
}

func TestMarshalUnmarshalGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	encodingConfig := simapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	am := gamm.NewAppModule(appCodec, app.GAMMKeeper, app.AccountKeeper, app.BankKeeper)
	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	app.BankKeeper.SetBalances(ctx, acc1, sdk.Coins{
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewInt64Coin("foo", 100000),
		sdk.NewInt64Coin("bar", 100000),
	})

	_, err := app.GAMMKeeper.CreatePool(ctx, acc1, types.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.NewDecWithPrec(1, 2),
	}, []types.PoolAsset{{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}, {
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}}, "")
	require.NoError(t, err)

	genesis := am.ExportGenesis(ctx, appCodec)
	assert.NotPanics(t, func() {
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})
		am := gamm.NewAppModule(appCodec, app.GAMMKeeper, app.AccountKeeper, app.BankKeeper)
		am.InitGenesis(ctx, appCodec, genesis)
	})
}
