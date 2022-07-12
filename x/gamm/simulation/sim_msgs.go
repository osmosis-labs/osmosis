package gammsimulation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	simulation "github.com/osmosis-labs/osmosis/v7/simulation/types"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
)

func SimulateJoinPoolMsg(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) {
	pool_id := simulation.RandLTBound(sim, k.GetNextPoolNumber(ctx))
	pool, err := k.GetPoolAndPoke(ctx, pool_id)
	poolDenoms := osmoutils.CoinsDenoms(pool.GetTotalPoolLiquidity(ctx))
	sender, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	_, _, _ = err, sender, senderExists
	// token_in_maxs := sim.FuzzTokensSubset(sender, k.GetPool(pool_id).Assets().Denoms())
	// share_out_amount := gamm.EstimateJoinPoolShareOut(ctx, pool_id, token_in_maxs)
	// share_out_amount = sim.FuzzEqualInt(share_out_amount)

	// return &gammtypes.MsgJoinPool{
	// 	sender,
	// 	pool_id,
	// 	token_in_maxs,
	// 	share_out_amount,
	// }
}
