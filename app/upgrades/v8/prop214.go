package v8

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammkeeper "github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
)

func Prop214(ctx sdk.Context, gamm *gammkeeper.Keeper) {
	poolId := 1
	pool, err := gamm.GetPoolAndPoke(ctx, uint64(poolId))
	if err != nil {
		panic(err)
	}

	balancerPool, ok := pool.(*balancer.Pool)
	if !ok {
		panic(ok)
	}

	balancerPool.PoolParams.SwapFee = sdk.MustNewDecFromStr("0.002")

	err = gamm.SetPool(ctx, balancerPool)
	if err != nil {
		panic(err)
	}
}
