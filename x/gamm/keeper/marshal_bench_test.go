package keeper_test

import (
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func genPoolAssets(r *rand.Rand) []gammtypes.PoolAsset {
	denoms := []string{"IBC/0123456789ABCDEF012346789ABCDEF", "IBC/denom56789ABCDEF012346789ABCDEF"}
	assets := []gammtypes.PoolAsset{}
	for _, denom := range denoms {
		amt, _ := simtypes.RandPositiveInt(r, sdk.NewIntWithDecimal(1, 40))
		reserveAmt := sdk.NewCoin(denom, amt)
		weight := sdk.NewInt(r.Int63n(9) + 1)
		assets = append(assets, gammtypes.PoolAsset{Token: reserveAmt, Weight: weight})
	}

	return assets
}

func genPoolParams(r *rand.Rand) balancer.PoolParams {
	swapFeeInt := int64(r.Intn(1e5))
	swapFee := sdk.NewDecWithPrec(swapFeeInt, 6)

	exitFeeInt := int64(r.Intn(1e5))
	exitFee := sdk.NewDecWithPrec(exitFeeInt, 6)

	// TODO: Randomly generate LBP params
	return balancer.PoolParams{
		SwapFee:                  swapFee,
		ExitFee:                  exitFee,
		SmoothWeightChangeParams: nil,
	}
}

func setupPools(maxNumPoolsToGen int) []gammtypes.PoolI {
	r := rand.New(rand.NewSource(10))
	// setup N pools
	pools := make([]gammtypes.PoolI, 0, maxNumPoolsToGen)
	for i := 0; i < maxNumPoolsToGen; i++ {
		assets := genPoolAssets(r)
		params := genPoolParams(r)
		pool, _ := balancer.NewBalancerPool(uint64(i), params, assets, "FutureGovernorString", time.Now())
		pools = append(pools, &pool)
	}
	return pools
}

func BenchmarkGammPoolSerialization(b *testing.B) {
	app := app.Setup(false)
	maxNumPoolsToGen := 5000
	pools := setupPools(maxNumPoolsToGen)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := i % maxNumPoolsToGen
		app.GAMMKeeper.MarshalPool(pools[j])
	}
}

func BenchmarkGammPoolDeserialization(b *testing.B) {
	app := app.Setup(false)
	maxNumPoolsToGen := 5000
	pools := setupPools(maxNumPoolsToGen)
	marshals := make([][]byte, 0, maxNumPoolsToGen)
	for i := 0; i < maxNumPoolsToGen; i++ {
		bz, _ := app.GAMMKeeper.MarshalPool(pools[i])
		marshals = append(marshals, bz)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := i % maxNumPoolsToGen
		app.GAMMKeeper.UnmarshalPool(marshals[j])
	}
}
