package gammsimulation

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	"github.com/osmosis-labs/osmosis/v12/simulation/simtypes"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

// RandomJoinPoolMsg pseudo-randomly selects an existing pool ID, attempts to find an account with the
// respective underlying token denoms, and attempts to execute a join pool transaction
func RandomJoinPoolMsg(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgJoinPool, error) {
	// get random pool
	pool_id, pool, _, _, poolDenoms, _, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// get address that has all denoms from the randomly selected pool
	sender, tokenIn, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denoms %s exists", poolDenoms)
	}

	// cap joining pool to the pool liquidity
	tokenIn = osmoutils.MinCoins(tokenIn, pool.GetTotalPoolLiquidity(ctx))

	// TODO: Fix API so this is a one liner, pool.CalcJoinPoolNoSwapShares()
	minShareOutAmt, err := deriveRealMinShareOutAmt(ctx, tokenIn, pool)
	if err != nil {
		return nil, err
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

// RandomExitPoolMsg pseudo-randomly selects an existing pool ID, attempts to find an account with the
// respective unbonded gamm shares, and attempts to execute an exit pool transaction
func RandomExitPoolMsg(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgExitPool, error) {
	// get random pool
	pool_id, pool, _, _, _, gammDenom, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// select an address that has gamm shares of the selected pool
	sender, gammShares, senderExists := sim.SelAddrWithDenom(ctx, gammDenom)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denom %s exists", gammDenom)
	}

	// calculate the minimum number of tokens received from input of gamm shares
	tokenOutMins, err := pool.CalcExitPoolCoinsFromShares(ctx, gammShares.Amount, pool.GetExitFee(ctx))
	if err != nil {
		return nil, err
	}

	return &types.MsgExitPool{
		Sender:        sender.Address.String(),
		PoolId:        pool_id,
		ShareInAmount: gammShares.Amount,
		TokenOutMins:  tokenOutMins,
	}, nil
}

// RandomJoinSwapExternAmountIn utilizes a random pool and with a random account does a single asset join with an exact
// amount of an asset for a minimum number of LP shares
func RandomJoinSwapExternAmountIn(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgJoinSwapExternAmountIn, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	pool_id, pool, coinIn, _, _, _, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// find an address with the coinIn denom and randomly select a subset of the coin
	sender, tokenIn, senderExists := sim.SelAddrWithDenom(ctx, coinIn.Denom)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denoms %s exists", coinIn.Denom)
	}

	// cap joining pool to the pool liquidity
	newTokenIn := osmoutils.MinCoins(sdk.NewCoins(coinIn), sdk.NewCoins(tokenIn))

	// calc shares out with tokenIn
	minShareOutAmt, _, err := pool.CalcJoinPoolShares(ctx, newTokenIn, pool.GetSwapFee(ctx))
	if err != nil {
		return nil, err
	}

	return &types.MsgJoinSwapExternAmountIn{
		Sender:            sender.Address.String(),
		PoolId:            pool_id,
		TokenIn:           newTokenIn[0],
		ShareOutMinAmount: minShareOutAmt,
	}, nil
}

// RandomJoinSwapShareAmountOut utilizes a random pool and with a random account and swaps a maximum of a specified token
// for an exact amount of LP shares
func RandomJoinSwapShareAmountOut(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgJoinSwapShareAmountOut, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	pool_id, pool, coinIn, _, _, _, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// find an address with the coinIn denom and randomly select a subset of the coin
	sender, tokenIn, senderExists := sim.SelAddrWithDenom(ctx, coinIn.Denom)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denoms %s exists", coinIn.Denom)
	}

	// cap joining pool to the pool liquidity
	newTokenIn := osmoutils.MinCoins(sdk.NewCoins(coinIn), sdk.NewCoins(tokenIn))

	// calc shares out with tokenIn
	minShareOutAmt, _, err := pool.CalcJoinPoolShares(ctx, newTokenIn, pool.GetSwapFee(ctx))
	if err != nil {
		return nil, err
	}

	// use PoolAmountOutExtension to calculate correct tokenInMaxAmount
	extendedPool, ok := pool.(types.PoolAmountOutExtension)
	if !ok {
		return nil, fmt.Errorf("pool with id %d does not support this kind of join", pool_id)
	}
	tokenInAmount, err := extendedPool.CalcTokenInShareAmountOut(ctx, tokenIn.Denom, minShareOutAmt, pool.GetSwapFee(ctx))
	if err != nil {
		return nil, err
	}

	return &types.MsgJoinSwapShareAmountOut{
		Sender:           sender.Address.String(),
		PoolId:           pool_id,
		TokenInDenom:     tokenIn.Denom,
		ShareOutAmount:   minShareOutAmt,
		TokenInMaxAmount: tokenInAmount,
	}, nil
}

// RandomExitSwapExternAmountOut utilizes a random pool and with a random account and swaps a maximum number of LP shares
// for an exact amount of one of the token pairs
func RandomExitSwapExternAmountOut(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgExitSwapExternAmountOut, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	pool_id, pool, coinIn, coinOut, _, gammDenom, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// select an address that has gamm shares of the selected pool
	sender, gammShares, senderExists := sim.SelAddrWithDenom(ctx, gammDenom)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denom %s exists", gammDenom)
	}

	// select a random amount of the account's gamm shares
	gammShares.Amount = sim.RandomAmount(gammShares.Amount)

	// calc exitedCoins from gammShares in
	exitedCoins, err := pool.CalcExitPoolCoinsFromShares(ctx, gammShares.Amount, pool.GetExitFee(ctx))
	if err != nil {
		return nil, err
	}

	// get amount of coinIn from exitedCoins and calculate how much of tokenOut you should get from that
	exitedCoinsIn := exitedCoins.AmountOf(coinIn.Denom)
	tokenOut, err := pool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(sdk.NewCoin(coinIn.Denom, exitedCoinsIn)), coinOut.Denom, pool.GetSwapFee(ctx))
	if err != nil {
		return nil, err
	}

	// TODO: figure out how to calculate the swap for the entire amount
	// I felt I was doing it correct but it was always off
	// since we are only doing half the swap out, we only use half the share in
	return &types.MsgExitSwapExternAmountOut{
		Sender:           sender.Address.String(),
		PoolId:           pool_id,
		TokenOut:         tokenOut,
		ShareInMaxAmount: gammShares.Amount.Quo(sdk.NewInt(2)),
	}, nil
}

// RandomExitSwapShareAmountIn utilizes a random pool and with a random account and swaps an number of LP shares
// for a minimum amount of one of the token pairs
func RandomExitSwapShareAmountIn(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgExitSwapShareAmountIn, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	pool_id, pool, coinIn, coinOut, _, gammDenom, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// select an address that has gamm shares of the selected pool
	sender, gammShares, senderExists := sim.SelAddrWithDenom(ctx, gammDenom)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denom %s exists", gammDenom)
	}

	// select a random amount of the account's gamm shares
	gammShares.Amount = sim.RandomAmount(gammShares.Amount)

	// calc exitedCoins from gammShares in
	exitedCoins, err := pool.CalcExitPoolCoinsFromShares(ctx, gammShares.Amount, pool.GetExitFee(ctx))
	if err != nil {
		return nil, err
	}

	// get amount of coinIn from exitedCoins and calculate how much of tokenOut you should get from that
	exitedCoinsIn := exitedCoins.AmountOf(coinIn.Denom)
	tokenOut, err := pool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(sdk.NewCoin(coinIn.Denom, exitedCoinsIn)), coinOut.Denom, pool.GetSwapFee(ctx))
	if err != nil {
		return nil, err
	}

	// TODO: figure out how to calculate the swap for the entire amount
	// I felt I was doing it correct but it was always off
	// since we are only doing half the swap out, we only use half the share in
	return &types.MsgExitSwapShareAmountIn{
		Sender:            sender.Address.String(),
		PoolId:            pool_id,
		TokenOutDenom:     tokenOut.Denom,
		ShareInAmount:     gammShares.Amount.Quo(sdk.NewInt(2)),
		TokenOutMinAmount: tokenOut.Amount.Quo(sdk.NewInt(2)),
	}, nil
}

// TODO: Fix CalcJoinPoolShares API so we don't have to do this
func deriveRealMinShareOutAmt(ctx sdk.Context, tokenIn sdk.Coins, pool types.TraditionalAmmInterface) (sdk.Int, error) {
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

func getRandPool(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (uint64, types.TraditionalAmmInterface, sdk.Coin, sdk.Coin, []string, string, error) {
	poolCount := k.GetPoolCount(ctx)
	if poolCount == 0 {
		return 0, nil, sdk.Coin{}, sdk.Coin{}, []string{}, "", errors.New("pool count is zero")
	}
	// select a pseudo-random pool ID, max bound by the upcoming pool ID
	pool_id := simtypes.RandLTBound(sim, poolCount)
	pool, err := k.GetPoolAndPoke(ctx, pool_id)
	if err != nil {
		return 0, nil, sdk.NewCoin("denom", sdk.ZeroInt()), sdk.NewCoin("denom", sdk.ZeroInt()), []string{}, "", err
	}
	poolCoins := pool.GetTotalPoolLiquidity(ctx)

	// TODO: Improve this, don't just assume two asset pools
	// randomly select one of the pool denoms to be the coinIn and one to be the coinOut
	r := sim.GetSeededRand("select random seed")
	index := r.Intn(len(poolCoins) - 1)
	coinIn := poolCoins[index]
	poolCoins = simtypes.RemoveIndex(poolCoins, index)
	coinOut := poolCoins[0]
	poolDenoms := osmoutils.CoinsDenoms(pool.GetTotalPoolLiquidity(ctx))
	gammDenom := types.GetPoolShareDenom(pool_id)
	return pool_id, pool, coinIn, coinOut, poolDenoms, gammDenom, err
}
