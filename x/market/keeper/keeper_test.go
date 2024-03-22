package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestOsmosisPoolDeltaUpdate(t *testing.T) {
	input := CreateTestInput(t)

	terraPoolDelta := input.MarketKeeper.GetOsmosisPoolDelta(input.Ctx)
	require.Equal(t, sdk.ZeroDec(), terraPoolDelta)

	diff := sdk.NewDec(10)
	input.MarketKeeper.SetOsmosisPoolDelta(input.Ctx, diff)

	terraPoolDelta = input.MarketKeeper.GetOsmosisPoolDelta(input.Ctx)
	require.Equal(t, diff, terraPoolDelta)
}

//// TestReplenishPools tests that
//// each pools move towards base pool
//func TestReplenishPools(t *testing.T) {
//	input := CreateTestInput(t)
//	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroSDRDenom, sdk.OneDec())
//
//	basePool := input.MarketKeeper.BasePool(input.Ctx)
//	terraPoolDelta := input.MarketKeeper.GetOsmosisPoolDelta(input.Ctx)
//	require.True(t, terraPoolDelta.IsZero())
//
//	// Positive delta
//	diff := basePool.QuoInt64((int64)(core.BlocksPerDay))
//	input.MarketKeeper.SetOsmosisPoolDelta(input.Ctx, diff)
//
//	input.MarketKeeper.ReplenishPools(input.Ctx)
//
//	terraPoolDelta = input.MarketKeeper.GetOsmosisPoolDelta(input.Ctx)
//	replenishAmt := diff.QuoInt64((int64)(input.MarketKeeper.PoolRecoveryPeriod(input.Ctx)))
//	expectedDelta := diff.Sub(replenishAmt)
//	require.Equal(t, expectedDelta, terraPoolDelta)
//
//	// Negative delta
//	diff = diff.Neg()
//	input.MarketKeeper.SetOsmosisPoolDelta(input.Ctx, diff)
//
//	input.MarketKeeper.ReplenishPools(input.Ctx)
//
//	osmosisPoolDelta = input.MarketKeeper.GetOsmosisPoolDelta(input.Ctx)
//	replenishAmt = diff.QuoInt64((int64)(input.MarketKeeper.PoolRecoveryPeriod(input.Ctx)))
//	expectedDelta = diff.Sub(replenishAmt)
//	require.Equal(t, expectedDelta, osmosisPoolDelta)
//}
