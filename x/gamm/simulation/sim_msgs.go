package gammsimulation

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	legacysimulationtype "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/osmosis-labs/osmosis/v10/osmoutils"
	"github.com/osmosis-labs/osmosis/v10/simulation/simtypes"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/keeper"
	balancertypes "github.com/osmosis-labs/osmosis/v10/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/types"
)

var (
	PoolCreationFee = sdk.NewInt64Coin("stake", 10_000_000)
)

// RandomJoinPoolMsg pseudo-randomly selects an existing pool ID, attempts to find an account with the
// respective underlying token denoms, and attempts to execute a join pool transaction
func RandomJoinPoolMsg(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgJoinPool, error) {
	// get random pool
	pool_id := simtypes.RandLTBound(sim, k.GetNextPoolNumber(ctx))
	pool, err := k.GetPoolAndPoke(ctx, pool_id)
	if err != nil {
		return &types.MsgJoinPool{}, err
	}
	// get address that has all denoms from the randomly selected pool
	poolDenoms := osmoutils.CoinsDenoms(pool.GetTotalPoolLiquidity(ctx))
	sender, tokenIn, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	if !senderExists {
		return &types.MsgJoinPool{}, fmt.Errorf("no sender with denoms %s exists", poolDenoms)
	}
	// cap joining pool to the pool liquidity
	tokenIn = osmoutils.MinCoins(tokenIn, pool.GetTotalPoolLiquidity(ctx))

	// TODO: Fix API so this is a one liner, pool.CalcJoinPoolNoSwapShares()
	minShareOutAmt, err := deriveRealMinShareOutAmt(ctx, tokenIn, pool)
	if err != nil {
		return &types.MsgJoinPool{}, err
	}

	// TODO: Make FuzzTokenSubset API, token_in_maxs := sim.FuzzTokensSubset(sender, poolDenoms)
	// TODO: Add some slippage tolerance
	// TODO: Make MinShareOutAmt fuzz API: minShareOutAmt = sim.FuzzEqualInt(share_out_amount)
	return &types.MsgJoinPool{
		Sender:         sender.Address.String(),
		PoolId:         pool_id,
		ShareOutAmount: minShareOutAmt,
		TokenInMaxs:    tokenIn,
	}, nil
}

// TODO: Fix CalcJoinPoolShares API so we don't have to do this
func deriveRealMinShareOutAmt(ctx sdk.Context, tokenIn sdk.Coins, pool types.PoolI) (sdk.Int, error) {
	minShareOutAmt, _, err := pool.CalcJoinPoolShares(ctx, tokenIn, pool.GetSwapFee(ctx))
	if err != nil {
		return sdk.Int{}, err
	}

	totalSharesAmount := pool.GetTotalShares()
	// shareRatio is the desired number of shares, divided by the total number of
	// shares currently in the pool. It is intended to be used in scenarios where you want
	shareRatio := minShareOutAmt.ToDec().QuoInt(totalSharesAmount)
	if shareRatio.LTE(sdk.ZeroDec()) {
		return sdk.Int{}, fmt.Errorf("share ratio is zero or negative")
	}

	poolLiquidity := pool.GetTotalPoolLiquidity(ctx)
	neededLpLiquidity := sdk.Coins{}

	for _, coin := range poolLiquidity {
		// (coin.Amt * shareRatio).Ceil()
		neededAmt := coin.Amount.ToDec().Mul(shareRatio).Ceil().RoundInt()
		if neededAmt.LTE(sdk.ZeroInt()) {
			return sdk.Int{}, fmt.Errorf("Too few shares out wanted")
		}
		neededCoin := sdk.Coin{Denom: coin.Denom, Amount: neededAmt}
		neededLpLiquidity = neededLpLiquidity.Add(neededCoin)
	}

	if tokenIn.Len() != 0 {
		if !(neededLpLiquidity.DenomsSubsetOf(tokenIn) && tokenIn.IsAllGTE(neededLpLiquidity)) {
			return sdk.Int{}, fmt.Errorf("TokenInMaxs is less than the needed LP liquidity to this JoinPoolNoSwap,"+
				" upperbound: %v, needed %v", tokenIn, neededLpLiquidity)
		} else if !(tokenIn.DenomsSubsetOf(neededLpLiquidity)) {
			return sdk.Int{}, fmt.Errorf("TokenInMaxs includes tokens that are not part of the target pool,"+
				" input tokens: %v, pool tokens %v", tokenIn, neededLpLiquidity)
		}
	}

	return minShareOutAmt, nil
}

// RandomExitPoolMsg pseudo-randomly selects an existing pool ID, attempts to find an account with the
// respective unbonded gamm shares, and attempts to execute an exit pool transaction
func RandomExitPoolMsg(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgExitPool, error) {
	// select a pseudo-random pool ID, max bound by the upcoming pool ID
	pool_id := simtypes.RandLTBound(sim, k.GetNextPoolNumber(ctx))
	pool, err := k.GetPoolAndPoke(ctx, pool_id)
	if err != nil {
		return &types.MsgExitPool{}, err
	}
	// select an address that has gamm shares of the selected pool
	gammDenom := types.GetPoolShareDenom(pool_id)
	sender, gammShares, senderExists := sim.SelAddrWithDenom(ctx, gammDenom)
	if !senderExists {
		return &types.MsgExitPool{}, fmt.Errorf("no sender with denom %s exists", gammDenom)
	}
	// calculate the minimum number of tokens received from input of gamm shares
	tokenOutMins, _ := pool.CalcExitPoolCoinsFromShares(ctx, gammShares.Amount, pool.GetExitFee(ctx))
	return &types.MsgExitPool{
		Sender:        sender.Address.String(),
		PoolId:        pool_id,
		ShareInAmount: gammShares.Amount,
		TokenOutMins:  tokenOutMins,
	}, nil
}

// RandomCreatePoolMsg attempts to find an account with two or more distinct denoms and attempts to send a
// create pool message composed of those denoms
func RandomCreateUniV2Msg(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*balancertypes.MsgCreateBalancerPool, error) {
	var poolAssets []balancertypes.PoolAsset
	// find an address with two or more distinct denoms in their wallet
	sender, senderExists := sim.RandomSimAccountWithConstraint(createPoolRestriction(k, sim, ctx))
	if !senderExists {
		return &balancertypes.MsgCreateBalancerPool{}, errors.New("no sender with two different denoms & pool creation fee exists")
	}
	poolCoins, _ := sim.GetRandSubsetOfKDenoms(ctx, sender, 2)
	if poolCoins.Add(PoolCreationFee).IsAnyGT(sim.BankKeeper().SpendableCoins(ctx, sender.Address)) {
		return &balancertypes.MsgCreateBalancerPool{}, errors.New("chose an account / creation amount that didn't pass fee bar")
	}

	// TODO: pseudo-randomly generate swap and exit fees
	poolParams := &balancertypes.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
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

func createPoolRestriction(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) simtypes.SimAccountConstraint {
	return func(acc legacysimulationtype.Account) bool {
		accCoins := sim.BankKeeper().SpendableCoins(ctx, acc.Address)
		hasTwoCoins := len(accCoins) >= 2
		hasPoolCreationFee := accCoins.AmountOf("stake").GT(PoolCreationFee.Amount)
		return hasTwoCoins && hasPoolCreationFee
	}
}
