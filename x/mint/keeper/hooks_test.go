package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/osmosis-labs/osmosis/app"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestEndOfEpochMintedCoinDistribution(t *testing.T) {
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

	params := app.IncentivesKeeper.GetParams(ctx)
	futureCtx := ctx.WithBlockTime(time.Now().Add(time.Minute))

	height := int64(1)
	lastHalvenPeriod := app.MintKeeper.GetLastHalvenEpochNum(ctx)
	// correct rewards
	for ; height < lastHalvenPeriod+app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs; height++ {
		feePoolOrigin := app.DistrKeeper.GetFeePool(ctx)
		app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
		app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

		// check community pool balance increase
		feePoolNew := app.DistrKeeper.GetFeePool(ctx)
		require.Equal(t, feePoolOrigin.CommunityPool.Add(sdk.NewDecCoin("stake", sdk.NewInt(2500000))), feePoolNew.CommunityPool, height)
	}

	app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
	app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

	lastHalvenPeriod = app.MintKeeper.GetLastHalvenEpochNum(ctx)
	require.Equal(t, lastHalvenPeriod, app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs)

	for ; height < lastHalvenPeriod+app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs; height++ {
		feePoolOrigin := app.DistrKeeper.GetFeePool(ctx)

		app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
		app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

		// check community pool balance increase with half reduction
		feePoolNew := app.DistrKeeper.GetFeePool(ctx)
		require.Equal(t, feePoolOrigin.CommunityPool.Add(sdk.NewDecCoin("stake", sdk.NewInt(1250000))), feePoolNew.CommunityPool)
	}
}

func TestEndOfEpochNoDistributionWhenIsNotYetStartTime(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	mintParams := app.MintKeeper.GetParams(ctx)
	mintParams.MintingRewardsDistributionStartTime = time.Now().Add(time.Hour * 100000)
	app.MintKeeper.SetParams(ctx, mintParams)

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

	params := app.IncentivesKeeper.GetParams(ctx)
	futureCtx := ctx.WithBlockTime(time.Now().Add(time.Minute))

	height := int64(1)
	lastHalvenPeriod := app.MintKeeper.GetLastHalvenEpochNum(ctx)
	// correct rewards
	for ; height < lastHalvenPeriod+app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs; height++ {
		feePoolOrigin := app.DistrKeeper.GetFeePool(ctx)
		app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
		app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

		// check community pool balance not increase
		feePoolNew := app.DistrKeeper.GetFeePool(ctx)
		require.Equal(t, feePoolOrigin.CommunityPool, feePoolNew.CommunityPool, height)
	}

	app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
	app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

	// halven period not updated as mint of tokens not going
	lastHalvenPeriod = app.MintKeeper.GetLastHalvenEpochNum(ctx)
	require.Equal(t, lastHalvenPeriod, int64(0))
}
