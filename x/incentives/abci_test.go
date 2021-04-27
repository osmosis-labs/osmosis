package incentives_test

import (
	"testing"
	"time"

	simapp "github.com/c-osmosis/osmosis/app"
	"github.com/c-osmosis/osmosis/x/incentives"
	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestPerpetualPotNotExpireAfterDistribution(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := sdk.AccAddress([]byte("addr1---------------"))

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	app.BankKeeper.SetBalances(ctx, addr, coins)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	_, err := app.IncentivesKeeper.CreatePot(ctx, true, addr, coins, distrTo, time.Now(), 1)
	require.NoError(t, err)

	_, beginBlock := app.IncentivesKeeper.GetCurrentEpochInfo(ctx)
	futureCtx := ctx.WithBlockHeight(beginBlock + app.IncentivesKeeper.GetParams(ctx).BlocksPerEpoch + 1)
	futureCtx = futureCtx.WithBlockTime(time.Now().Add(time.Minute))
	incentives.EndBlocker(futureCtx, app.IncentivesKeeper)
	pots := app.IncentivesKeeper.GetUpcomingPots(futureCtx)
	require.Len(t, pots, 0)
	pots = app.IncentivesKeeper.GetActivePots(futureCtx)
	require.Len(t, pots, 1)
	pots = app.IncentivesKeeper.GetFinishedPots(futureCtx)
	require.Len(t, pots, 0)
}

func TestNonPerpetualPotExpireAfterDistribution(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := sdk.AccAddress([]byte("addr1---------------"))

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	app.BankKeeper.SetBalances(ctx, addr, coins)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	_, err := app.IncentivesKeeper.CreatePot(ctx, false, addr, coins, distrTo, time.Now(), 1)
	require.NoError(t, err)

	_, beginBlock := app.IncentivesKeeper.GetCurrentEpochInfo(ctx)
	futureCtx := ctx.WithBlockHeight(beginBlock + app.IncentivesKeeper.GetParams(ctx).BlocksPerEpoch + 1)
	futureCtx = futureCtx.WithBlockTime(time.Now().Add(time.Minute))
	incentives.EndBlocker(futureCtx, app.IncentivesKeeper)
	pots := app.IncentivesKeeper.GetUpcomingPots(futureCtx)
	require.Len(t, pots, 0)
	pots = app.IncentivesKeeper.GetActivePots(futureCtx)
	require.Len(t, pots, 0)
	pots = app.IncentivesKeeper.GetFinishedPots(futureCtx)
	require.Len(t, pots, 1)
}
