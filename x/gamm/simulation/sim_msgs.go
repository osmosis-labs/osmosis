package gammsimulation

// import (
// 	sdk "github.com/cosmos/cosmos-sdk/types"

// 	simulation "github.com/osmosis-labs/osmosis/v7/simulation/types"
// 	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
// )

// func SimulateJoinPoolMsg(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) sdk.Msg {
// 	pool_id := sim.RandLessThan(k.GetNextPoolId(ctx))
// 	sender := sim.FuzzAddrWithDenoms(k.GetPool(ctx, pool_id).Assets())
// 	token_in_maxs := sim.FuzzTokensSubset(sender, k.GetPool(pool_id).Assets().Denoms())
// 	share_out_amount := gamm.EstimateJoinPoolShareOut(ctx, pool_id, token_in_maxs)
// 	share_out_amount = sim.FuzzEqualInt(share_out_amount)

// 	return simtypes.NewOperationFromMsg(
// 		&MsgJoinPool{
// 			sender,
// 			pool_id,
// 			token_in_maxs,
// 			share_out_amount,
// 		})
// }
