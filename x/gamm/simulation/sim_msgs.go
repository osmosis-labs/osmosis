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
	// get random pool
	pool_id := simulation.RandLTBound(sim, k.GetNextPoolNumber(ctx))
	pool, err := k.GetPoolAndPoke(ctx, pool_id)
	if err != nil {
		return &gammtypes.MsgJoinPool{}, err
	}
	// get address that has all denoms from the randomly selected pool
	poolDenoms := osmoutils.CoinsDenoms(pool.GetTotalPoolLiquidity(ctx))
	sender, tokenIn, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	if !senderExists {
		return &gammtypes.MsgJoinPool{}, fmt.Errorf("no sender with denoms %s exists", poolDenoms)
	}

	// cap joining pool to double the pool liquidity
	for i, coinPool := range pool.GetTotalPoolLiquidity(ctx) {
		if coinPool.Amount.LT(tokenIn[i].Amount.Quo(sdk.NewInt(2))) {
			return &gammtypes.MsgJoinPool{}, fmt.Errorf("join pool is capped to double the pool liquidity")
		}
	}

	minShareOutAmt, _, err := pool.CalcJoinPoolShares(ctx, tokenIn, pool.GetSwapFee(ctx))
	if err != nil {
		return &gammtypes.MsgJoinPool{}, err
	}

	totalSharesAmount := pool.GetTotalShares()
	// shareRatio is the desired number of shares, divided by the total number of
	// shares currently in the pool. It is intended to be used in scenarios where you want
	shareRatio := minShareOutAmt.ToDec().QuoInt(totalSharesAmount)
	if shareRatio.LTE(sdk.ZeroDec()) {
		return &gammtypes.MsgJoinPool{}, fmt.Errorf("share ratio is zero or negative")
	}

	poolLiquidity := pool.GetTotalPoolLiquidity(ctx)
	neededLpLiquidity := sdk.Coins{}

	for _, coin := range poolLiquidity {
		// (coin.Amt * shareRatio).Ceil()
		neededAmt := coin.Amount.ToDec().Mul(shareRatio).Ceil().RoundInt()
		if neededAmt.LTE(sdk.ZeroInt()) {
			return &gammtypes.MsgJoinPool{}, fmt.Errorf("Too few shares out wanted")
		}
		neededCoin := sdk.Coin{Denom: coin.Denom, Amount: neededAmt}
		neededLpLiquidity = neededLpLiquidity.Add(neededCoin)
	}

	if tokenIn.Len() != 0 {
		if !(neededLpLiquidity.DenomsSubsetOf(tokenIn) && tokenIn.IsAllGTE(neededLpLiquidity)) {
			return &gammtypes.MsgJoinPool{}, fmt.Errorf("TokenInMaxs is less than the needed LP liquidity to this JoinPoolNoSwap,"+
				" upperbound: %v, needed %v", tokenIn, neededLpLiquidity)
		} else if !(tokenIn.DenomsSubsetOf(neededLpLiquidity)) {
			return &gammtypes.MsgJoinPool{}, fmt.Errorf("TokenInMaxs includes tokens that are not part of the target pool,"+
				" input tokens: %v, pool tokens %v", tokenIn, neededLpLiquidity)
		}
	}

	// TODO: Make FuzzTokenSubset API, token_in_maxs := sim.FuzzTokensSubset(sender, poolDenoms)
	// TODO: Add some slippage tolerance
	// TODO: Make MinShareOutAmt fuzz API: minShareOutAmt = sim.FuzzEqualInt(share_out_amount)
	return &gammtypes.MsgJoinPool{
		Sender:         sender.Address.String(),
		PoolId:         pool_id,
		ShareOutAmount: minShareOutAmt,
		TokenInMaxs:    tokenIn,
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
		return &balancertypes.MsgCreateBalancerPool{}, fmt.Errorf("provided sender %s does not posses two unique denoms %s", sender.Address.String(), poolCoins.String())
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
