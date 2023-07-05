package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

// CalculateSpotPrice returns the spot price of the quote asset in terms of the base asset,
// using the specified pool.
// E.g. if pool 1 trades 2 atom for 3 osmo, the quote asset was atom, and the base asset was osmo,
// this would return 1.5. (Meaning that 1 atom costs 1.5 osmo)
//
// This function is guaranteed to not panic, but may return an error if:
// * An internal error within the pool occurs for calculating the spot price
// * The returned spot price is greater than max spot price
func (k Keeper) CalculateSpotPrice(
	ctx sdk.Context,
	poolID uint64,
	quoteAssetDenom string,
	baseAssetDenom string,
) (spotPrice sdk.Dec, err error) {
	pool, err := k.GetPoolAndPoke(ctx, poolID)
	if err != nil {
		return sdk.Dec{}, err
	}

	// defer to catch panics, in case something internal overflows.
	defer func() {
		if r := recover(); r != nil {
			spotPrice = sdk.Dec{}
			err = types.ErrSpotPriceInternal
		}
	}()

	spotPrice, err = pool.SpotPrice(ctx, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	// if spotPrice greater than max spot price, return an error
	if spotPrice.GT(types.MaxSpotPrice) {
		return types.MaxSpotPrice, types.ErrSpotPriceOverflow
	} else if !spotPrice.IsPositive() {
		return sdk.Dec{}, types.ErrSpotPriceInternal
	}

	// we want to round this to `SpotPriceSigFigs` of precision
	spotPrice = osmomath.SigFigRound(spotPrice, types.SpotPriceSigFigs)
	return spotPrice, err
}

// This function:
// - saves the pool to state
// - Mints LP shares to the pool creator
// - Sets bank metadata for the LP denom
// - Records total liquidity increase
// - Calls the AfterPoolCreated hook
func (k Keeper) InitializePool(ctx sdk.Context, pool poolmanagertypes.PoolI, sender sdk.AccAddress) (err error) {
	cfmmPool, err := asCFMMPool(pool)
	if err != nil {
		return err
	}

	exitFee := cfmmPool.GetExitFee(ctx)
	if !exitFee.IsZero() {
		return fmt.Errorf("can not create pool with non zero exit fee, got %d", exitFee)
	}

	// Mint the initial pool shares share token to the sender
	err = k.MintPoolShareToAccount(ctx, pool, sender, cfmmPool.GetTotalShares())
	if err != nil {
		return err
	}

	// Finally, add the share token's meta data to the bank keeper.
	poolShareBaseDenom := types.GetPoolShareDenom(pool.GetId())
	poolShareDisplayDenom := fmt.Sprintf("GAMM-%d", pool.GetId())
	k.bankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
		Description: fmt.Sprintf("The share token of the gamm pool %d", pool.GetId()),
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    poolShareBaseDenom,
				Exponent: 0,
				Aliases: []string{
					"attopoolshare",
				},
			},
			{
				Denom:    poolShareDisplayDenom,
				Exponent: types.OneShareExponent,
				Aliases:  nil,
			},
		},
		Base:    poolShareBaseDenom,
		Display: poolShareDisplayDenom,
	})

	if err := k.setPool(ctx, pool); err != nil {
		return err
	}

	// N.B.: these hooks propagate to x/twap to create
	// twap records at pool creation time.
	// Additionally, these hooks are used in x/pool-incentives to
	// create gauges.
	k.hooks.AfterCFMMPoolCreated(ctx, sender, pool.GetId())
	k.RecordTotalLiquidityIncrease(ctx, cfmmPool.GetTotalPoolLiquidity(ctx))
	return nil
}

// JoinPoolNoSwap aims to LP exactly enough to pool #{poolId} to get shareOutAmount number of LP shares.
// If the required tokens is greater than tokenInMaxs, returns an error & the message reverts.
// Leftover tokens that weren't LP'd (due to being at inexact ratios) remain in the sender account.
//
// JoinPoolNoSwap determines the maximum amount that can be LP'd without any swap,
// by looking at the ratio of the total LP'd assets. (e.g. 2 osmo : 1 atom)
// It then finds the maximal amount that can be LP'd.
func (k Keeper) JoinPoolNoSwap(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	shareOutAmount sdk.Int,
	tokenInMaxs sdk.Coins,
) (tokenIn sdk.Coins, sharesOut sdk.Int, err error) {
	// defer to catch panics, in case something internal overflows.
	defer func() {
		if r := recover(); r != nil {
			tokenIn = sdk.Coins{}
			sharesOut = sdk.Int{}
			err = fmt.Errorf("function JoinPoolNoSwap failed due to internal reason: %v", r)
		}
	}()
	// all pools handled within this method are pointer references, `JoinPool` directly updates the pools
	pool, err := k.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return nil, sdk.ZeroInt(), err
	}

	// we do an abstract calculation on the lp liquidity coins needed to have
	// the designated amount of given shares of the pool without performing swap
	neededLpLiquidity, err := getMaximalNoSwapLPAmount(ctx, pool, shareOutAmount)
	if err != nil {
		return nil, sdk.ZeroInt(), err
	}

	// check that needed lp liquidity does not exceed the given `tokenInMaxs` parameter. Return error if so.
	// if tokenInMaxs == 0, don't do this check.
	if tokenInMaxs.Len() != 0 {
		if !(neededLpLiquidity.DenomsSubsetOf(tokenInMaxs)) {
			return nil, sdk.ZeroInt(), errorsmod.Wrapf(types.ErrLimitMaxAmount, "TokenInMaxs does not include all the tokens that are part of the target pool,"+
				" upperbound: %v, needed %v", tokenInMaxs, neededLpLiquidity)
		} else if !(tokenInMaxs.DenomsSubsetOf(neededLpLiquidity)) {
			return nil, sdk.ZeroInt(), errorsmod.Wrapf(types.ErrDenomNotFoundInPool, "TokenInMaxs includes tokens that are not part of the target pool,"+
				" input tokens: %v, pool tokens %v", tokenInMaxs, neededLpLiquidity)
		}
		if !(tokenInMaxs.IsAllGTE(neededLpLiquidity)) {
			return nil, sdk.ZeroInt(), errorsmod.Wrapf(types.ErrLimitMaxAmount, "TokenInMaxs is less than the needed LP liquidity to this JoinPoolNoSwap,"+
				" upperbound: %v, needed %v", tokenInMaxs, neededLpLiquidity)
		}
	}

	sharesOut, err = pool.JoinPoolNoSwap(ctx, neededLpLiquidity, pool.GetSpreadFactor(ctx))
	if err != nil {
		return nil, sdk.ZeroInt(), err
	}
	// sanity check, don't return error as not worth halting the LP. We know its not too much.
	if sharesOut.LT(shareOutAmount) {
		ctx.Logger().Error(fmt.Sprintf("Expected to JoinPoolNoSwap >= %s shares, actually did %s shares",
			shareOutAmount, sharesOut))
	}

	err = k.applyJoinPoolStateChange(ctx, pool, sender, sharesOut, neededLpLiquidity)
	return neededLpLiquidity, sharesOut, err
}

// getMaximalNoSwapLPAmount returns the coins(lp liquidity) needed to get the specified amount of shares in the pool.
// Steps to getting the needed lp liquidity coins needed for the share of the pools are
// 1. calculate how much percent of the pool does given share account for(# of input shares / # of current total shares)
// 2. since we know how much % of the pool we want, iterate through all pool liquidity to calculate how much coins we need for
// each pool asset.
func getMaximalNoSwapLPAmount(ctx sdk.Context, pool types.CFMMPoolI, shareOutAmount sdk.Int) (neededLpLiquidity sdk.Coins, err error) {
	totalSharesAmount := pool.GetTotalShares()
	// shareRatio is the desired number of shares, divided by the total number of
	// shares currently in the pool. It is intended to be used in scenarios where you want
	shareRatio := shareOutAmount.ToDec().QuoInt(totalSharesAmount)
	if shareRatio.LTE(sdk.ZeroDec()) {
		return sdk.Coins{}, errorsmod.Wrapf(types.ErrInvalidMathApprox, "Too few shares out wanted. "+
			"(debug: getMaximalNoSwapLPAmount share ratio is zero or negative)")
	}

	poolLiquidity := pool.GetTotalPoolLiquidity(ctx)
	neededLpLiquidity = sdk.Coins{}

	for _, coin := range poolLiquidity {
		// (coin.Amt * shareRatio).Ceil()
		neededAmt := coin.Amount.ToDec().Mul(shareRatio).Ceil().RoundInt()
		if neededAmt.LTE(sdk.ZeroInt()) {
			return sdk.Coins{}, errorsmod.Wrapf(types.ErrInvalidMathApprox, "Too few shares out wanted")
		}
		neededCoin := sdk.Coin{Denom: coin.Denom, Amount: neededAmt}
		neededLpLiquidity = neededLpLiquidity.Add(neededCoin)
	}
	return neededLpLiquidity, nil
}

// JoinSwapExactAmountIn is an LP transaction, that will LP all of the provided
// tokensIn coins. The underlying pool is responsible for swapping any non-even
// LP proportions to the correct ratios. An error is returned if the amount of
// LP shares obtained at the end is less than shareOutMinAmount. Otherwise, we
// return the total amount of shares outgoing from joining the pool.
func (k Keeper) JoinSwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokensIn sdk.Coins,
	shareOutMinAmount sdk.Int,
) (sharesOut sdk.Int, err error) {
	// defer to catch panics, in case something internal overflows.
	defer func() {
		if r := recover(); r != nil {
			sharesOut = sdk.Int{}
			err = fmt.Errorf("function JoinSwapExactAmountIn failed due to internal reason: %v", r)
		}
	}()

	pool, err := k.GetCFMMPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	sharesOut, err = pool.JoinPool(ctx, tokensIn, pool.GetSpreadFactor(ctx))
	switch {
	case err != nil:
		return sdk.ZeroInt(), err

	case sharesOut.LT(shareOutMinAmount):
		return sdk.ZeroInt(), errorsmod.Wrapf(
			types.ErrLimitMinAmount,
			"too much slippage; needed a minimum of %s shares to pass, got %s",
			shareOutMinAmount, sharesOut,
		)

	case sharesOut.LTE(sdk.ZeroInt()):
		return sdk.ZeroInt(), errorsmod.Wrapf(types.ErrInvalidMathApprox, "share amount is zero or negative")
	}

	if err := k.applyJoinPoolStateChange(ctx, pool, sender, sharesOut, tokensIn); err != nil {
		return sdk.ZeroInt(), err
	}

	return sharesOut, nil
}

func (k Keeper) JoinSwapShareAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenInDenom string,
	shareOutAmount sdk.Int,
	tokenInMaxAmount sdk.Int,
) (tokenInAmount sdk.Int, err error) {
	// defer to catch panics, in case something internal overflows.
	defer func() {
		if r := recover(); r != nil {
			tokenInAmount = sdk.Int{}
			err = fmt.Errorf("function JoinSwapShareAmountOut failed due to internal reason: %v", r)
		}
	}()

	pool, err := k.GetCFMMPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	extendedPool, ok := pool.(types.PoolAmountOutExtension)
	if !ok {
		return sdk.Int{}, fmt.Errorf("pool with id %d does not support this kind of join", poolId)
	}

	tokenInAmount, err = extendedPool.CalcTokenInShareAmountOut(ctx, tokenInDenom, shareOutAmount, pool.GetSpreadFactor(ctx))
	if err != nil {
		return sdk.Int{}, err
	}

	if tokenInAmount.GT(tokenInMaxAmount) {
		return sdk.Int{}, errorsmod.Wrapf(types.ErrLimitMaxAmount, "%s resulted tokens is larger than the max amount of %s", tokenInAmount, tokenInMaxAmount)
	}

	tokenIn := sdk.NewCoins(sdk.NewCoin(tokenInDenom, tokenInAmount))
	// Not using generic JoinPool because we want to guarantee exact shares out
	extendedPool.IncreaseLiquidity(shareOutAmount, tokenIn)

	err = k.applyJoinPoolStateChange(ctx, pool, sender, shareOutAmount, tokenIn)
	if err != nil {
		return sdk.ZeroInt(), err
	}
	return tokenInAmount, nil
}

func (k Keeper) ExitPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	shareInAmount sdk.Int,
	tokenOutMins sdk.Coins,
) (exitCoins sdk.Coins, err error) {
	pool, err := k.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return sdk.Coins{}, err
	}

	totalSharesAmount := pool.GetTotalShares()
	if shareInAmount.GTE(totalSharesAmount) {
		return sdk.Coins{}, errorsmod.Wrapf(types.ErrInvalidMathApprox, "Trying to exit >= the number of shares contained in the pool.")
	} else if shareInAmount.LTE(sdk.ZeroInt()) {
		return sdk.Coins{}, errorsmod.Wrapf(types.ErrInvalidMathApprox, "Trying to exit a negative amount of shares")
	}
	exitFee := pool.GetExitFee(ctx)
	exitCoins, err = pool.ExitPool(ctx, shareInAmount, exitFee)
	if err != nil {
		return sdk.Coins{}, err
	}
	if !tokenOutMins.DenomsSubsetOf(exitCoins) || tokenOutMins.IsAnyGT(exitCoins) {
		return sdk.Coins{}, errorsmod.Wrapf(types.ErrLimitMinAmount,
			"Exit pool returned %s , minimum tokens out specified as %s",
			exitCoins, tokenOutMins)
	}

	err = k.applyExitPoolStateChange(ctx, pool, sender, shareInAmount, exitCoins)
	if err != nil {
		return sdk.Coins{}, err
	}

	return exitCoins, nil
}

// ExitSwapShareAmountIn is an Exit Pool transaction, that will exit all of the provided LP shares,
// and then swap it all against the pool into tokenOutDenom.
// If the amount of tokens gotten out after the swap is less than tokenOutMinAmount, return an error.
func (k Keeper) ExitSwapShareAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenOutDenom string,
	shareInAmount sdk.Int,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	exitCoins, err := k.ExitPool(ctx, sender, poolId, shareInAmount, sdk.Coins{})
	if err != nil {
		return sdk.Int{}, err
	}

	pool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}
	spreadFactor := pool.GetSpreadFactor(ctx)

	tokenOutAmount = exitCoins.AmountOf(tokenOutDenom)
	for _, coin := range exitCoins {
		if coin.Denom == tokenOutDenom {
			continue
		}
		swapOut, err := k.SwapExactAmountIn(ctx, sender, pool, coin, tokenOutDenom, sdk.ZeroInt(), spreadFactor)
		if err != nil {
			return sdk.Int{}, err
		}
		tokenOutAmount = tokenOutAmount.Add(swapOut)
	}

	if tokenOutAmount.LT(tokenOutMinAmount) {
		return sdk.Int{}, errorsmod.Wrapf(types.ErrLimitMinAmount,
			"Provided LP shares yield %s tokens out, wanted a minimum of %s for it to work",
			tokenOutAmount, tokenOutMinAmount)
	}
	return tokenOutAmount, nil
}

func (k Keeper) ExitSwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenOut sdk.Coin,
	shareInMaxAmount sdk.Int,
) (shareInAmount sdk.Int, err error) {
	pool, err := k.GetCFMMPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	extendedPool, ok := pool.(types.PoolAmountOutExtension)
	if !ok {
		return sdk.Int{}, fmt.Errorf("pool with id %d does not support this kind of exit", poolId)
	}

	shareInAmount, err = extendedPool.ExitSwapExactAmountOut(ctx, tokenOut, shareInMaxAmount)
	if err != nil {
		return sdk.Int{}, err
	}

	if err := k.applyExitPoolStateChange(ctx, pool, sender, shareInAmount, sdk.Coins{tokenOut}); err != nil {
		return sdk.Int{}, err
	}

	return shareInAmount, nil
}
