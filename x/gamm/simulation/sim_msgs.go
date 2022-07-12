package gammsimulation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	simulation "github.com/osmosis-labs/osmosis/v7/simulation/types"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func SimulateJoinPoolMsg(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) *gammtypes.MsgJoinPool {
	pool_id := simulation.RandLTBound(sim, k.GetNextPoolNumber(ctx))
	pool, err := k.GetPoolAndPoke(ctx, pool_id)
	poolDenoms := osmoutils.CoinsDenoms(pool.GetTotalPoolLiquidity(ctx))
	sender, tokenInMaxs, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	_, _, _, _ = err, sender, tokenInMaxs, senderExists
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
