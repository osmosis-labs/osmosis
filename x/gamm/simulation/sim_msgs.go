package gammsimulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/server/rosetta/lib/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	simulation "github.com/osmosis-labs/osmosis/v7/simulation/types"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	balancertypes "github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func CurrySimMsgJoinPool(k keeper.Keeper) func(sim *simulation.SimCtx, ctx sdk.Context) (*gammtypes.MsgJoinPool, error) {
	return func(sim *simulation.SimCtx, ctx sdk.Context) (*gammtypes.MsgJoinPool, error) {
		return RandomJoinPoolMsg(k, sim, ctx)
	}
}

func RandomJoinPoolMsg(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) (*gammtypes.MsgJoinPool, error) {
	// Get pool
	pool_id := simulation.RandLTBound(sim, k.GetNextPoolNumber(ctx))
	pool, err := k.GetPoolAndPoke(ctx, pool_id)
	if err != nil {
		return &gammtypes.MsgJoinPool{}, err
	}
	// Get address that has all denoms in the pool
	poolDenoms := osmoutils.CoinsDenoms(pool.GetTotalPoolLiquidity(ctx))
	sender, tokenInMaxs, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	if !senderExists {
		return &gammtypes.MsgJoinPool{}, fmt.Errorf("no sender with denoms %s exists", poolDenoms)
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
	}, nil
}

func RandomCreateUniv2PoolMsg(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) (*balancertypes.MsgCreateBalancerPool, error) {
	// 1) Select two denoms, ideally with some frequency weighting based on distribution amongst addrs
	// 2) Select sender with both denoms + creation fee
	// 3) Create pool
	return nil, errors.ErrNotImplemented
}
