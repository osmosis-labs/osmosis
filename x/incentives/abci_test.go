package incentives_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	osmoapp "github.com/osmosis-labs/osmosis/v8/app"
	lockuptypes "github.com/osmosis-labs/osmosis/v8/x/lockup/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestPerpetualGaugeNotExpireAfterDistribution(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := sdk.AccAddress([]byte("addr1---------------"))

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	err := simapp.FundAccount(app.BankKeeper, ctx, addr, coins)
	require.NoError(t, err)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}

	// mints coins so supply exists on chain
	mintLPtokens := sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	err = simapp.FundAccount(app.BankKeeper, ctx, addr, mintLPtokens)
	require.NoError(t, err)

	_, err = app.IncentivesKeeper.CreateGauge(ctx, true, addr, coins, distrTo, time.Now(), 1)
	require.NoError(t, err)

	params := app.IncentivesKeeper.GetParams(ctx)
	futureCtx := ctx.WithBlockTime(time.Now().Add(time.Minute))
	app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, 1)
	app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, 1)
	gauges := app.IncentivesKeeper.GetUpcomingGauges(futureCtx)
	require.Len(t, gauges, 0)
	gauges = app.IncentivesKeeper.GetActiveGauges(futureCtx)
	require.Len(t, gauges, 1)
	gauges = app.IncentivesKeeper.GetFinishedGauges(futureCtx)
	require.Len(t, gauges, 0)
}

func TestNonPerpetualGaugeExpireAfterDistribution(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := sdk.AccAddress([]byte("addr1---------------"))

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	err := simapp.FundAccount(app.BankKeeper, ctx, addr, coins)
	require.NoError(t, err)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}

	// mints coins so supply exists on chain
	mintLPtokens := sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	err = simapp.FundAccount(app.BankKeeper, ctx, addr, mintLPtokens)
	require.NoError(t, err)

	_, err = app.IncentivesKeeper.CreateGauge(ctx, false, addr, coins, distrTo, time.Now(), 1)
	require.NoError(t, err)

	params := app.IncentivesKeeper.GetParams(ctx)
	futureCtx := ctx.WithBlockTime(time.Now().Add(time.Minute))
	app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, 1)
	app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, 1)
	gauges := app.IncentivesKeeper.GetUpcomingGauges(futureCtx)
	require.Len(t, gauges, 0)
	gauges = app.IncentivesKeeper.GetActiveGauges(futureCtx)
	require.Len(t, gauges, 0)
	gauges = app.IncentivesKeeper.GetFinishedGauges(futureCtx)
	require.Len(t, gauges, 1)
}
