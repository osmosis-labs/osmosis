package market

import (
	"testing"

	"github.com/osmosis-labs/osmosis/v23/x/market/keeper"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestReplenishPools(t *testing.T) {
	input := keeper.CreateTestInput(t)

	osmosisDelta := sdk.NewDecWithPrec(17987573223725367, 3)
	input.MarketKeeper.SetOsmosisPoolDelta(input.Ctx, osmosisDelta)

	for i := 0; i < 100; i++ {
		osmosisDelta = input.MarketKeeper.GetOsmosisPoolDelta(input.Ctx)

		poolRecoveryPeriod := int64(input.MarketKeeper.PoolRecoveryPeriod(input.Ctx))
		osmosisRegressionAmt := osmosisDelta.QuoInt64(poolRecoveryPeriod)

		EndBlocker(input.Ctx, input.MarketKeeper)

		osmosisPoolDelta := input.MarketKeeper.GetOsmosisPoolDelta(input.Ctx)
		require.Equal(t, osmosisDelta.Sub(osmosisRegressionAmt), osmosisPoolDelta)
	}
}
