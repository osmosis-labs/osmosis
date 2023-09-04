package v9

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	gammkeeper "github.com/osmosis-labs/osmosis/v19/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v19/x/gamm/pool-models/balancer"
)

// Executes prop214, https://www.mintscan.io/osmosis/proposals/214
// Run `osmosisd q gov proposal 214` to see the text.
// It was voted in, and it has update instructions:
// Voting YES for this proposal would reduce the Pool 1 (OSMO/ATOM) spread factor from 0.3% to 0.2%
func ExecuteProp214(ctx sdk.Context, gamm *gammkeeper.Keeper) {
	poolId := 1
	pool, err := gamm.GetPoolAndPoke(ctx, uint64(poolId))
	if err != nil {
		panic(err)
	}

	balancerPool, ok := pool.(*balancer.Pool)
	if !ok {
		panic(ok)
	}

	balancerPool.PoolParams.SwapFee = osmomath.MustNewDecFromStr("0.002")

	// Kept as comments for recordkeeping. SetPool is now private:
	// 		err = gamm.SetPool(ctx, balancerPool)
	// 		if err != nil {
	//	 		panic(err)
	//  	}
}
