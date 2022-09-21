package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

// EstimateMultihopSwapExactAmountIn iterates `estimateSwapExactAmountIn` and returns
// the total token out given route for multihop.
func (k Keeper) EstimateMultihopSwapExactAmountIn(
	ctx sdk.Context,
	routes []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	// use cache context so that pool state is not mutated after estimation
	cacheCtx, _ := ctx.CacheContext()
	for i, route := range routes {
		swapFeeMultiplier := sdk.OneDec()
		if types.SwapAmountInRoutes(routes).IsOsmoRoutedMultihop() {
			swapFeeMultiplier = types.MultihopSwapFeeMultiplierForOsmoPools.Clone()
		}

		// To prevent the multihop swap from being interrupted prematurely, we keep
		// the minimum expected output at a very low number until the last pool
		_outMinAmount := sdk.NewInt(1)
		if len(routes)-1 == i {
			_outMinAmount = tokenOutMinAmount
		}

		// Execute the expected swap on the current routed pool
		pool, poolErr := k.getPoolForSwap(cacheCtx, route.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}

		swapFee := pool.GetSwapFee(cacheCtx).Mul(swapFeeMultiplier)

		// Execute the expected swap on the current routed pool
		// call internal method with useCacheCtx = false to prevent double caching,
		// we need to save pool state for the next iteration of pool route
		tokenOutAmount, err = k.estimateSwapExactAmountIn(cacheCtx, route.PoolId, tokenIn, route.TokenOutDenom, _outMinAmount, swapFee, false)
		if err != nil {
			return sdk.Int{}, err
		}

		// Chain output of current pool as the input for the next routed pool
		tokenIn = sdk.NewCoin(route.TokenOutDenom, tokenOutAmount)
	}
	return tokenOutAmount, nil
}

// EstimateSwapExactAmountIn estimates the amount of token out given the exact amount of token in.
// This method does not execute the full steps of an actaul swap, nor does it change the state after estimation.
// Returns the estimated amount of token out using cache context.
func (k Keeper) EstimateSwapExactAmountIn(
	ctx sdk.Context,
	poolId uint64,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
	swapFee sdk.Dec,
) (sdk.Int, error) {
	return k.estimateSwapExactAmountIn(ctx, poolId, tokenIn, tokenOutDenom, tokenOutMinAmount, swapFee, true)
}

// estimateSwapExactAmountIn returns the amount of token out given the amount of token in.
// This method should only be called by estimate methods, as this method does not include full steps of an actual swap.
func (k Keeper) estimateSwapExactAmountIn(
	ctx sdk.Context,
	poolId uint64,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
	swapFee sdk.Dec,
	useCacheCtx bool,
) (sdk.Int, error) {
	if useCacheCtx {
		ctx, _ = ctx.CacheContext()
	}
	pool, err := k.getPoolForSwap(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	tokenOut, err := k.swapExactAmountInNoTokenSend(ctx, pool, tokenIn, tokenOutDenom, tokenOutMinAmount, swapFee)
	if err != nil {
		return tokenOut.Amount, err
	}

	err = k.updatePoolForSwap(ctx, pool, tokenIn, tokenOut)
	return tokenOut.Amount, err
}

// EstimateMultihopSwapExactAmountOut iterates `estimateSwapExactAmountOut` and returns
// the total token in given route for multihop.
func (k Keeper) EstimateMultihopSwapExactAmountOut(
	ctx sdk.Context,
	routes []types.SwapAmountOutRoute,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	// use cache context so that pool state is not mutated after estimation
	cacheCtx, _ := ctx.CacheContext()
	swapFeeMultiplier := sdk.OneDec()

	if types.SwapAmountOutRoutes(routes).IsOsmoRoutedMultihop() {
		swapFeeMultiplier = types.MultihopSwapFeeMultiplierForOsmoPools.Clone()
	}

	// Determine what the estimated input would be for each pool along the multihop route
	insExpected, err := k.createMultihopExpectedSwapOuts(cacheCtx, routes, tokenOut, swapFeeMultiplier)
	if err != nil {
		return sdk.Int{}, err
	}
	if len(insExpected) == 0 {
		return sdk.Int{}, nil
	}

	insExpected[0] = tokenInMaxAmount

	// Iterates through each routed pool and executes their respective swaps. Note that all of the work to get the return
	// value of this method is done when we calculate insExpected – this for loop primarily serves to execute the actual
	// swaps on each pool.
	for i, route := range routes {
		_tokenOut := tokenOut

		// If there is one pool left in the route, set the expected output of the current swap
		// to the estimated input of the final pool.
		if i != len(routes)-1 {
			_tokenOut = sdk.NewCoin(routes[i+1].TokenInDenom, insExpected[i+1])
		}

		// Execute the expected swap on the current routed pool
		pool, poolErr := k.getPoolForSwap(cacheCtx, route.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}
		swapFee := pool.GetSwapFee(cacheCtx).Mul(swapFeeMultiplier)
		// Execute the expected swap on the current routed pool
		// call internal method with useCacheCtx = false to prevent double caching,
		// we need to save pool state for the next iteration of pool route
		_tokenInAmount, err := k.estimateSwapExactAmountOut(cacheCtx, route.PoolId, route.TokenInDenom, insExpected[i], _tokenOut, swapFee, false)
		if err != nil {
			return sdk.Int{}, err
		}

		// Sets the final amount of tokens that need to be input into the first pool. Even though this is the final return value for the
		// whole method and will not change after the first iteration, we still iterate through the rest of the pools to execute their respective
		// swaps.
		if i == 0 {
			tokenInAmount = _tokenInAmount
		}
	}

	return tokenInAmount, nil
}

// EstimateSwapExactAmountOut estimates the amount of token in given the exact amount of token out.
// This method does not execute the full steps of an actaul swap, nor does it change the state after estimation.
// Returns the estimated amount of token out using cache context.
func (k Keeper) EstimateSwapExactAmountOut(
	ctx sdk.Context,
	poolId uint64,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
	swapFee sdk.Dec,
) (tokenInAmount sdk.Int, err error) {
	return k.estimateSwapExactAmountOut(ctx, poolId, tokenInDenom, tokenInMaxAmount, tokenOut, swapFee, true)
}

// estimateSwapExactAmountOut returns the amount of token in given the amount of token out.
// This method should only be called by estimate methods, as this method does not include full steps of an actual swap.
func (k Keeper) estimateSwapExactAmountOut(
	ctx sdk.Context,
	poolId uint64,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
	swapFee sdk.Dec,
	useCacheCtx bool,
) (tokenInAmount sdk.Int, err error) {
	if useCacheCtx {
		ctx, _ = ctx.CacheContext()
	}
	pool, err := k.getPoolForSwap(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	tokenIn, err := k.swapExactAmountOutNoTokenSend(ctx, pool, tokenInDenom, tokenInMaxAmount, tokenOut, swapFee)
	if err != nil {
		return tokenIn.Amount, err
	}

	err = k.updatePoolForSwap(ctx, pool, tokenIn, tokenOut)
	return tokenIn.Amount, err
}
