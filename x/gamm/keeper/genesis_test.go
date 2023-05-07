package keeper_test

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	osmoapp "github.com/osmosis-labs/osmosis/v15/app"
	"github.com/osmosis-labs/osmosis/v15/x/gamm"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
)

var (
	DefaultMigrationRecords = types.MigrationRecords{BalancerToConcentratedPoolLinks: []types.BalancerToConcentratedPoolLink{
		{BalancerPoolId: 1, ClPoolId: 4},
		{BalancerPoolId: 2, ClPoolId: 5},
		{BalancerPoolId: 3, ClPoolId: 6},
	}}
)

func TestGammInitGenesis(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	balancerPool, err := balancer.NewBalancerPool(1, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}, []balancer.PoolAsset{
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

	any, err := codectypes.NewAnyWithValue(&balancerPool)
	require.NoError(t, err)

	app.GAMMKeeper.InitGenesis(ctx, types.GenesisState{
		Pools:          []*codectypes.Any{any},
		NextPoolNumber: 2,
		Params: types.Params{
			PoolCreationFee: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000_000_000)},
		},
		MigrationRecords: &DefaultMigrationRecords,
	}, app.AppCodec())

	require.Equal(t, app.PoolManagerKeeper.GetNextPoolId(ctx), uint64(1))
	poolStored, err := app.GAMMKeeper.GetPoolAndPoke(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, balancerPool.GetId(), poolStored.GetId())
	require.Equal(t, balancerPool.GetAddress(), poolStored.GetAddress())
	require.Equal(t, balancerPool.GetSwapFee(ctx), poolStored.GetSwapFee(ctx))
	require.Equal(t, balancerPool.GetExitFee(ctx), poolStored.GetExitFee(ctx))
	// require.Equal(t, balancerPool.GetTotalWeight(), sdk.Nw)
	require.Equal(t, balancerPool.GetTotalShares(), poolStored.GetTotalShares())
	// require.Equal(t, balancerPool.GetAllPoolAssets(), poolStored.GetAllPoolAssets())
	require.Equal(t, balancerPool.String(), poolStored.String())

	_, err = app.GAMMKeeper.GetPoolAndPoke(ctx, 2)
	require.Error(t, err)

	liquidity := app.GAMMKeeper.GetTotalLiquidity(ctx)
	require.Equal(t, liquidity, sdk.Coins{sdk.NewInt64Coin("nodetoken", 10), sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)})

	postInitGenMigrationRecords, err := app.GAMMKeeper.GetAllMigrationInfo(ctx)
	require.NoError(t, err)
	require.Equal(t, DefaultMigrationRecords, postInitGenMigrationRecords)
}

func TestGammExportGenesis(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	err := simapp.FundAccount(app.BankKeeper, ctx, acc1, sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewInt64Coin("foo", 100000),
		sdk.NewInt64Coin("bar", 100000),
	))
	require.NoError(t, err)

	msg := balancer.NewMsgCreateBalancerPool(acc1, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}, []balancer.PoolAsset{{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}, {
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}}, "")
	_, err = app.PoolManagerKeeper.CreatePool(ctx, msg)
	require.NoError(t, err)

	msg = balancer.NewMsgCreateBalancerPool(acc1, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}, []balancer.PoolAsset{{
		Weight: sdk.NewInt(70),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}, {
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}}, "")
	_, err = app.PoolManagerKeeper.CreatePool(ctx, msg)
	require.NoError(t, err)

	app.GAMMKeeper.OverwriteMigrationRecords(ctx, DefaultMigrationRecords)

	genesis := app.GAMMKeeper.ExportGenesis(ctx)
	// Note: the next pool number index has been migrated to
	// poolmanager.
	// The reason it is kept in gamm is for migrations.
	// As a result, it is 1 here. This index is to be removed
	// in a subsequent upgrade.
	require.Equal(t, genesis.NextPoolNumber, uint64(1))
	require.Len(t, genesis.Pools, 2)
	require.Equal(t, genesis.MigrationRecords, &DefaultMigrationRecords)
}

func TestMarshalUnmarshalGenesis(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	encodingConfig := osmoapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	am := gamm.NewAppModule(appCodec, *app.GAMMKeeper, app.AccountKeeper, app.BankKeeper)
	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	err := simapp.FundAccount(app.BankKeeper, ctx, acc1, sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewInt64Coin("foo", 100000),
		sdk.NewInt64Coin("bar", 100000),
	))
	require.NoError(t, err)

	msg := balancer.NewMsgCreateBalancerPool(acc1, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}, []balancer.PoolAsset{{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}, {
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}}, "")
	_, err = app.PoolManagerKeeper.CreatePool(ctx, msg)
	require.NoError(t, err)

	app.GAMMKeeper.OverwriteMigrationRecords(ctx, DefaultMigrationRecords)

	genesis := am.ExportGenesis(ctx, appCodec)
	assert.NotPanics(t, func() {
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})
		am := gamm.NewAppModule(appCodec, *app.GAMMKeeper, app.AccountKeeper, app.BankKeeper)
		am.InitGenesis(ctx, appCodec, genesis)
	})
}
