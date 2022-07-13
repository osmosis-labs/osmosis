package gammsimulation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	simulation "github.com/osmosis-labs/osmosis/v7/simulation/types"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func CurrySimMsgJoinPool(k keeper.Keeper) func(sim *simulation.SimCtx, ctx sdk.Context) *gammtypes.MsgJoinPool {
	return func(sim *simulation.SimCtx, ctx sdk.Context) *gammtypes.MsgJoinPool {
		return SimulateJoinPoolMsg(k, sim, ctx)
	}
}

func SimulateJoinPoolMsg(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) *gammtypes.MsgJoinPool {
	pool_id := simulation.RandLTBound(sim, k.GetNextPoolNumber(ctx))
	pool, err := k.GetPoolAndPoke(ctx, pool_id)
	if err != nil {
		return &gammtypes.MsgJoinPool{}
	}
	poolDenoms := osmoutils.CoinsDenoms(pool.GetTotalPoolLiquidity(ctx))
	sender, tokenInMaxs, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	if !senderExists {
		return &gammtypes.MsgJoinPool{}
	}
	// TODO: Make FuzzTokenSubset API, token_in_maxs := sim.FuzzTokensSubset(sender, poolDenoms)
	// TODO: Add some slippage tolerance
	minShareOutAmt, _, _ := pool.CalcJoinPoolShares(ctx, tokenInMaxs, pool.GetSwapFee(ctx))
	// TODO: Make MinShareOutAmt fuzz API: minShareOutAmt = sim.FuzzEqualInt(share_out_amount)
	return &gammtypes.MsgJoinPool{
		Sender:         sender.Address.String(),
		PoolId:         pool_id,
		ShareOutAmount: minShareOutAmt,
		TokenInMaxs:    tokenInMaxs,
	}
}
