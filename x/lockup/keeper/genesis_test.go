package keeper_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	osmoapp "github.com/osmosis-labs/osmosis/v27/app"
	"github.com/osmosis-labs/osmosis/v27/x/lockup"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
)

var (
	now         = time.Now().UTC()
	acc1        = sdk.AccAddress([]byte("addr1---------------"))
	acc2        = sdk.AccAddress([]byte("addr2---------------"))
	testGenesis = types.GenesisState{
		LastLockId: 10,
		Locks: []types.PeriodLock{
			{
				ID:                    1,
				Owner:                 acc1.String(),
				RewardReceiverAddress: "",
				Duration:              time.Second,
				EndTime:               time.Time{},
				Coins:                 sdk.Coins{sdk.NewInt64Coin("foo", 10000000)},
			},
			{
				ID:                    2,
				Owner:                 acc1.String(),
				RewardReceiverAddress: acc2.String(),
				Duration:              time.Hour,
				EndTime:               time.Time{},
				Coins:                 sdk.Coins{sdk.NewInt64Coin("foo", 15000000)},
			},
			{
				ID:                    3,
				Owner:                 acc2.String(),
				RewardReceiverAddress: acc1.String(),
				Duration:              time.Minute,
				EndTime:               time.Time{},
				Coins:                 sdk.Coins{sdk.NewInt64Coin("foo", 5000000)},
			},
		},
		Params: &types.Params{
			ForceUnlockAllowedAddresses: []string{acc1.String(), acc2.String()},
		},
	}
)

func TestInitGenesis(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	app.LockupKeeper.InitGenesis(ctx, genesis)

	coins := app.LockupKeeper.GetAccountLockedCoins(ctx, acc1)
	require.Equal(t, coins.String(), sdk.NewInt64Coin("foo", 25000000).String())

	coins = app.LockupKeeper.GetAccountLockedCoins(ctx, acc2)
	require.Equal(t, coins.String(), sdk.NewInt64Coin("foo", 5000000).String())

	lastLockId := app.LockupKeeper.GetLastLockID(ctx)
	require.Equal(t, lastLockId, uint64(10))

	acc := app.LockupKeeper.GetPeriodLocksAccumulation(ctx, types.QueryCondition{
		Denom:    "foo",
		Duration: time.Second,
	})
	require.Equal(t, osmomath.NewInt(30000000), acc)

	params := app.LockupKeeper.GetParams(ctx)
	require.Equal(t, params.ForceUnlockAllowedAddresses, []string{acc1.String(), acc2.String()})
}

func TestExportGenesis(t *testing.T) {
	dirName := fmt.Sprintf("%d", rand.Int())
	app := osmoapp.SetupWithCustomHome(false, dirName)

	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	app.LockupKeeper.InitGenesis(ctx, genesis)

	err := testutil.FundAccount(ctx, app.BankKeeper, acc2, sdk.Coins{sdk.NewInt64Coin("foo", 5000000)})
	require.NoError(t, err)
	_, err = app.LockupKeeper.CreateLock(ctx, acc2, sdk.Coins{sdk.NewInt64Coin("foo", 5000000)}, time.Second*5)
	require.NoError(t, err)

	coins := app.LockupKeeper.GetAccountLockedCoins(ctx, acc2)
	require.Equal(t, coins.String(), sdk.NewInt64Coin("foo", 10000000).String())

	genesisExported := app.LockupKeeper.ExportGenesis(ctx)
	require.Equal(t, genesisExported.LastLockId, uint64(11))
	require.Equal(t, genesisExported.Locks, []types.PeriodLock{
		{
			ID:                    1,
			Owner:                 acc1.String(),
			RewardReceiverAddress: "",
			Duration:              time.Second,
			EndTime:               time.Time{},
			Coins:                 sdk.Coins{sdk.NewInt64Coin("foo", 10000000)},
		},
		{
			ID:                    11,
			Owner:                 acc2.String(),
			RewardReceiverAddress: "",
			Duration:              time.Second * 5,
			EndTime:               time.Time{},
			Coins:                 sdk.Coins{sdk.NewInt64Coin("foo", 5000000)},
		},
		{
			ID:                    3,
			Owner:                 acc2.String(),
			RewardReceiverAddress: acc1.String(),
			Duration:              time.Minute,
			EndTime:               time.Time{},
			Coins:                 sdk.Coins{sdk.NewInt64Coin("foo", 5000000)},
		},
		{
			ID:                    2,
			Owner:                 acc1.String(),
			RewardReceiverAddress: acc2.String(),
			Duration:              time.Hour,
			EndTime:               time.Time{},
			Coins:                 sdk.Coins{sdk.NewInt64Coin("foo", 15000000)},
		},
	})
	require.Equal(t, genesisExported.Params, &types.Params{
		ForceUnlockAllowedAddresses: []string{acc1.String(), acc2.String()},
	})
	os.RemoveAll(dirName)
}

func TestMarshalUnmarshalGenesis(t *testing.T) {
	dirName := fmt.Sprintf("%d", rand.Int())
	app := osmoapp.SetupWithCustomHome(false, dirName)

	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	encodingConfig := osmoapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	am := lockup.NewAppModule(*app.LockupKeeper, app.AccountKeeper, app.BankKeeper)

	err := testutil.FundAccount(ctx, app.BankKeeper, acc2, sdk.Coins{sdk.NewInt64Coin("foo", 5000000)})
	require.NoError(t, err)
	_, err = app.LockupKeeper.CreateLock(ctx, acc2, sdk.Coins{sdk.NewInt64Coin("foo", 5000000)}, time.Second*5)
	require.NoError(t, err)

	genesisExported := am.ExportGenesis(ctx, appCodec)
	os.RemoveAll(dirName)
	assert.NotPanics(t, func() {
		app := osmoapp.Setup(false)
		ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := lockup.NewAppModule(*app.LockupKeeper, app.AccountKeeper, app.BankKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}
