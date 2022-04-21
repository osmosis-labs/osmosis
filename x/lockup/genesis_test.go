package lockup_test

import (
	"testing"
	"time"

	osmoapp "github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/x/lockup"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	now         = time.Now().UTC()
	acc1        = sdk.AccAddress([]byte("addr1---------------"))
	acc2        = sdk.AccAddress([]byte("addr2---------------"))
	testGenesis = types.GenesisState{
		LastLockId: 10,
		Locks: []types.PeriodLock{
			{
				ID:       1,
				Owner:    acc1.String(),
				Duration: time.Second,
				EndTime:  time.Time{},
				Coins:    sdk.Coins{sdk.NewInt64Coin("foo", 10000000)},
			},
			{
				ID:       2,
				Owner:    acc1.String(),
				Duration: time.Hour,
				EndTime:  time.Time{},
				Coins:    sdk.Coins{sdk.NewInt64Coin("foo", 15000000)},
			},
			{
				ID:       3,
				Owner:    acc2.String(),
				Duration: time.Minute,
				EndTime:  time.Time{},
				Coins:    sdk.Coins{sdk.NewInt64Coin("foo", 5000000)},
			},
		},
	}
)

func TestInitGenesis(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	lockup.InitGenesis(ctx, *app.LockupKeeper, genesis)

	coins := app.LockupKeeper.GetAccountLockedCoins(ctx, acc1)
	require.Equal(t, coins.String(), sdk.NewInt64Coin("foo", 25000000).String())

	coins = app.LockupKeeper.GetAccountLockedCoins(ctx, acc2)
	require.Equal(t, coins.String(), sdk.NewInt64Coin("foo", 5000000).String())

	// TODO: module account balance is kept by bank keeper and no need to check here
	// coins = app.LockupKeeper.GetModuleBalance(ctx)
	// require.Equal(t, coins.String(), sdk.NewInt64Coin("foo", 30000000).String())

	lastLockId := app.LockupKeeper.GetLastLockID(ctx)
	require.Equal(t, lastLockId, uint64(10))
}

func TestExportGenesis(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	lockup.InitGenesis(ctx, *app.LockupKeeper, genesis)

	err := simapp.FundAccount(app.BankKeeper, ctx, acc2, sdk.Coins{sdk.NewInt64Coin("foo", 5000000)})
	require.NoError(t, err)
	_, err = app.LockupKeeper.LockTokens(ctx, acc2, sdk.Coins{sdk.NewInt64Coin("foo", 5000000)}, time.Second*5)
	require.NoError(t, err)

	coins := app.LockupKeeper.GetAccountLockedCoins(ctx, acc2)
	require.Equal(t, coins.String(), sdk.NewInt64Coin("foo", 10000000).String())

	genesisExported := lockup.ExportGenesis(ctx, *app.LockupKeeper)
	require.Equal(t, genesisExported.LastLockId, uint64(11))
	require.Equal(t, genesisExported.Locks, []types.PeriodLock{
		{
			ID:       1,
			Owner:    acc1.String(),
			Duration: time.Second,
			EndTime:  time.Time{},
			Coins:    sdk.Coins{sdk.NewInt64Coin("foo", 10000000)},
		},
		{
			ID:       11,
			Owner:    acc2.String(),
			Duration: time.Second * 5,
			EndTime:  time.Time{},
			Coins:    sdk.Coins{sdk.NewInt64Coin("foo", 5000000)},
		},
		{
			ID:       3,
			Owner:    acc2.String(),
			Duration: time.Minute,
			EndTime:  time.Time{},
			Coins:    sdk.Coins{sdk.NewInt64Coin("foo", 5000000)},
		},
		{
			ID:       2,
			Owner:    acc1.String(),
			Duration: time.Hour,
			EndTime:  time.Time{},
			Coins:    sdk.Coins{sdk.NewInt64Coin("foo", 15000000)},
		},
	})
}

func TestMarshalUnmarshalGenesis(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	encodingConfig := osmoapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	am := lockup.NewAppModule(appCodec, *app.LockupKeeper, app.AccountKeeper, app.BankKeeper)

	err := simapp.FundAccount(app.BankKeeper, ctx, acc2, sdk.Coins{sdk.NewInt64Coin("foo", 5000000)})
	require.NoError(t, err)
	_, err = app.LockupKeeper.LockTokens(ctx, acc2, sdk.Coins{sdk.NewInt64Coin("foo", 5000000)}, time.Second*5)
	require.NoError(t, err)

	genesisExported := am.ExportGenesis(ctx, appCodec)
	assert.NotPanics(t, func() {
		app := osmoapp.Setup(false)
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := lockup.NewAppModule(appCodec, *app.LockupKeeper, app.AccountKeeper, app.BankKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}
