package gammsimulation

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	simulation "github.com/osmosis-labs/osmosis/v7/simulation/types"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	balancertypes "github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

// RandomJoinPoolMsg pseudo-randomly selects an existing pool ID, attempts to find an account with the
// respective underlying token denoms, and attempts to execute a join pool transaction
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

// RandomExitPoolMsg pseudo-randomly selects an existing pool ID, attempts to find an account with the
// respective unbonded gamm shares, and attempts to execute an exit pool transaction
func RandomExitPoolMsg(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) (*gammtypes.MsgExitPool, error) {
	// select a pseudo-random pool ID, max bound by the upcoming pool ID
	pool_id := simulation.RandLTBound(sim, k.GetNextPoolNumber(ctx))
	pool, err := k.GetPoolAndPoke(ctx, pool_id)
	if err != nil {
		return &gammtypes.MsgExitPool{}, err
	}
	// select an address that has gamm shares of the selected pool
	gammDenom := fmt.Sprintf("gamm/pool/%v", pool_id)
	sender, gammShares, senderExists := sim.SelAddrWithDenom(ctx, gammDenom)
	if !senderExists {
		return &gammtypes.MsgExitPool{}, fmt.Errorf("no sender with denom %s exists", gammDenom)
	}
	// calculate the minimum number of tokens received from input of gamm shares
	tokenOutMins, _ := pool.CalcExitPoolCoinsFromShares(ctx, gammShares.Amount, pool.GetExitFee(ctx))
	return &gammtypes.MsgExitPool{
		Sender:        sender.Address.String(),
		PoolId:        pool_id,
		ShareInAmount: gammShares.Amount,
		TokenOutMins:  tokenOutMins,
	}, nil
}

// RandomCreatePoolMsg attempts to find an account with two or more distinct denoms and attempts to send a
// create pool message composed of those denoms
func RandomCreateUniV2Msg(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) (*balancertypes.MsgCreateBalancerPool, error) {
	var poolAssets []balancertypes.PoolAsset
	// find an address with two or more distinct denoms in their wallet
	sender, senderExists := sim.RandomSimAccountWithKDenoms(ctx, 2)
	if !senderExists {
		return &balancertypes.MsgCreateBalancerPool{}, fmt.Errorf("no sender with two different denoms exists")
	}
	poolCoins, denomsExist := sim.GetRandSubsetOfKDenoms(ctx, sender, 2)
	if !denomsExist {
		return &balancertypes.MsgCreateBalancerPool{}, fmt.Errorf("provided sender does not posses two unique denoms")
	}
	// TODO: pseudo-randomly generate swap and exit fees
	poolParams := &balancertypes.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.NewDecWithPrec(1, 2),
	}

	// from the above selected account, determine the token type and respective weight needed to make the pool
	for i := 0; i < len(poolCoins); i++ {
		poolAssets = append(poolAssets, balancertypes.PoolAsset{
			Weight: sdk.OneInt(),
			Token:  poolCoins[i],
		})
	}
	return &balancertypes.MsgCreateBalancerPool{
		Sender:     sender.Address.String(),
		PoolParams: poolParams,
		PoolAssets: poolAssets,
	}, nil
}
