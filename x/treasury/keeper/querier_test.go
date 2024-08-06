package keeper

import (
	"testing"

	core "github.com/osmosis-labs/osmosis/v23/types"
	"github.com/osmosis-labs/osmosis/v23/x/treasury/types"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

func TestQueryParams(t *testing.T) {
	input := CreateTestInput(t)
	ctx := sdk.WrapSDKContext(input.Ctx)

	querier := NewQuerier(input.TreasuryKeeper)
	res, err := querier.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)

	require.Equal(t, input.TreasuryKeeper.GetParams(input.Ctx), res.Params)
}

func TestQueryRewardWeight(t *testing.T) {
	input := CreateTestInput(t)
	ctx := sdk.WrapSDKContext(input.Ctx)

	querier := NewQuerier(input.TreasuryKeeper)
	res, err := querier.RewardWeight(ctx, &types.QueryRewardWeightRequest{})
	require.NoError(t, err)

	require.Equal(t, input.TreasuryKeeper.GetRewardWeight(input.Ctx), res.RewardWeight)
}

func TestQueryTaxRate(t *testing.T) {
	input := CreateTestInput(t)
	ctx := sdk.WrapSDKContext(input.Ctx)

	querier := NewQuerier(input.TreasuryKeeper)
	res, err := querier.TaxRate(ctx, &types.QueryTaxRateRequest{})
	require.NoError(t, err)

	require.Equal(t, input.TreasuryKeeper.GetTaxRate(input.Ctx), res.TaxRate)
}

func TestQueryTaxCap(t *testing.T) {
	input := CreateTestInput(t)
	ctx := sdk.WrapSDKContext(input.Ctx)

	taxCap := sdk.NewInt(1000000000)
	input.TreasuryKeeper.SetTaxCap(input.Ctx, core.MicroKRWDenom, taxCap)

	querier := NewQuerier(input.TreasuryKeeper)
	res, err := querier.TaxCap(ctx, &types.QueryTaxCapRequest{
		Denom: core.MicroKRWDenom,
	})
	require.NoError(t, err)

	require.Equal(t, taxCap, res.TaxCap)
	require.Equal(t, input.TreasuryKeeper.GetTaxCap(input.Ctx, core.MicroKRWDenom), res.TaxCap)
}

func TestQueryTaxCaps(t *testing.T) {
	input := CreateTestInput(t)
	ctx := sdk.WrapSDKContext(input.Ctx)

	input.TreasuryKeeper.SetTaxCap(input.Ctx, "ukrw", sdk.NewInt(1000000000))
	input.TreasuryKeeper.SetTaxCap(input.Ctx, "usdr", sdk.NewInt(1000000))
	input.TreasuryKeeper.SetTaxCap(input.Ctx, "uusd", sdk.NewInt(1200000))

	// Get a currency super random; should default to policy coin.
	querier := NewQuerier(input.TreasuryKeeper)
	res, err := querier.TaxCaps(ctx, &types.QueryTaxCapsRequest{})
	require.NoError(t, err)

	require.Equal(t, []types.QueryTaxCapsResponseItem{{
		Denom:  "ukrw",
		TaxCap: sdk.NewInt(1000000000),
	}, {
		Denom:  "usdr",
		TaxCap: sdk.NewInt(1000000),
	}, {
		Denom:  "uusd",
		TaxCap: sdk.NewInt(1200000),
	}}, res.TaxCaps)
}

func TestQueryTaxProceeds(t *testing.T) {
	input := CreateTestInput(t)
	ctx := sdk.WrapSDKContext(input.Ctx)

	taxProceeds := sdk.NewCoins(sdk.NewCoin(core.MicroLunaDenom, sdk.NewInt(100)), sdk.NewCoin(core.MicroKRWDenom, sdk.NewInt(100)))
	input.TreasuryKeeper.SetEpochTaxProceeds(input.Ctx, taxProceeds)

	querier := NewQuerier(input.TreasuryKeeper)
	res, err := querier.TaxProceeds(ctx, &types.QueryTaxProceedsRequest{})
	require.NoError(t, err)

	require.Equal(t, taxProceeds, res.TaxProceeds)
	require.Equal(t, input.TreasuryKeeper.PeekEpochTaxProceeds(input.Ctx), res.TaxProceeds)
}

func TestQuerySeigniorageProceeds(t *testing.T) {
	input := CreateTestInput(t)
	ctx := sdk.WrapSDKContext(input.Ctx)

	targetSeigniorage := sdk.NewInt(10)

	input.TreasuryKeeper.RecordEpochInitialIssuance(input.Ctx)

	input.Ctx = input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek))
	err := input.BankKeeper.BurnCoins(input.Ctx, faucetAccountName, sdk.NewCoins(sdk.NewCoin(core.MicroLunaDenom, targetSeigniorage)))
	require.NoError(t, err)

	querier := NewQuerier(input.TreasuryKeeper)
	res, err := querier.SeigniorageProceeds(ctx, &types.QuerySeigniorageProceedsRequest{})
	require.NoError(t, err)

	require.Equal(t, targetSeigniorage, res.SeigniorageProceeds)
	require.Equal(t, input.TreasuryKeeper.PeekEpochSeigniorage(input.Ctx), res.SeigniorageProceeds)
}

func TestQueryIndicators(t *testing.T) {
	input := CreateTestInput(t)
	ctx := sdk.WrapSDKContext(input.Ctx)

	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(input.StakingKeeper)

	stakingAmt := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
	addr, val := ValAddrs[0], ValPubKeys[0]
	addr1, val1 := ValAddrs[1], ValPubKeys[1]
	_, err := stakingMsgSvr.CreateValidator(input.Ctx, NewTestMsgCreateValidator(addr, val, stakingAmt))
	require.NoError(t, err)
	_, err = stakingMsgSvr.CreateValidator(input.Ctx, NewTestMsgCreateValidator(addr1, val1, stakingAmt))
	require.NoError(t, err)

	staking.EndBlocker(input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek)-1), input.StakingKeeper)

	proceedsAmt := sdk.NewInt(1000000000000)
	taxProceeds := sdk.NewCoins(sdk.NewCoin(core.MicroSDRDenom, proceedsAmt))
	input.TreasuryKeeper.RecordEpochTaxProceeds(input.Ctx, taxProceeds)

	targetIndicators := &types.QueryIndicatorsResponse{
		TRLYear:  sdk.NewDecFromInt(proceedsAmt).QuoInt(stakingAmt.MulRaw(2)),
		TRLMonth: sdk.NewDecFromInt(proceedsAmt).QuoInt(stakingAmt.MulRaw(2)),
	}

	querier := NewQuerier(input.TreasuryKeeper)
	res, err := querier.Indicators(ctx, &types.QueryIndicatorsRequest{})
	require.NoError(t, err)
	require.Equal(t, targetIndicators, res)

	// Update indicators
	input.TreasuryKeeper.UpdateIndicators(input.Ctx)

	// Record same tax proceeds to get same trl
	input.TreasuryKeeper.RecordEpochTaxProceeds(input.Ctx, taxProceeds)

	// Change context to next epoch
	input.Ctx = input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek))
	res, err = querier.Indicators(ctx, &types.QueryIndicatorsRequest{})
	require.NoError(t, err)
	require.Equal(t, targetIndicators, res)
}
