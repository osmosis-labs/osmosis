package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, spotPriceAfter sdk.Dec, err error) {
	if tokenIn.Denom == tokenOutDenom {
		return sdk.Int{}, sdk.Dec{}, errors.New("cannot trade same denomination in and out")
	}

	pool, inPoolAsset, outPoolAsset, err :=
		k.getPoolAndInOutAssets(ctx, poolId, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	if !pool.IsActive(ctx.BlockTime()) {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrPoolLocked, "swap on inactive pool")
	}

	// TODO: Understand if we are handling swap fee consistently,
	// with the global swap fee and the pool swap fee

	tokenOutAmount = calcOutGivenIn(
		inPoolAsset.Token.Amount.ToDec(),
		inPoolAsset.Weight.ToDec(),
		outPoolAsset.Token.Amount.ToDec(),
		outPoolAsset.Weight.ToDec(),
		tokenIn.Amount.ToDec(),
		pool.GetPoolSwapFee(),
	).TruncateInt()
	if tokenOutAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenOutAmount.LT(tokenOutMinAmount) {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrLimitMinAmount, "%s token is lesser than min amount", outPoolAsset.Token.Denom)
	}

	inPoolAsset.Token.Amount = inPoolAsset.Token.Amount.Add(tokenIn.Amount)
	outPoolAsset.Token.Amount = outPoolAsset.Token.Amount.Sub(tokenOutAmount)

	tokenOut := sdk.Coin{Denom: tokenOutDenom, Amount: tokenOutAmount}

	err = k.updatePoolForSwap(ctx, pool, sender, inPoolAsset, outPoolAsset, tokenIn, tokenOut)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	return tokenOutAmount, spotPriceAfter, nil
}

func (k Keeper) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, spotPriceAfter sdk.Dec, err error) {
	if tokenInDenom == tokenOut.Denom {
		return sdk.Int{}, sdk.Dec{}, errors.New("cannot trade same denomination in and out")
	}

	pool, inPoolAsset, outPoolAsset, err :=
		k.getPoolAndInOutAssets(ctx, poolId, tokenInDenom, tokenOut.Denom)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	if !pool.IsActive(ctx.BlockTime()) {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrPoolLocked, "swap on inactive pool")
	}

	poolOutBal, _ := pool.GetTokenBalance(tokenOut.Denom)
	if tokenOut.Amount.GTE(poolOutBal) {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrTooManyTokensOut,
			"can't get more tokens out than there are tokens in the pool")
	}

	tokenInAmount = calcInGivenOut(
		inPoolAsset.Token.Amount.ToDec(),
		inPoolAsset.Weight.ToDec(),
		outPoolAsset.Token.Amount.ToDec(),
		outPoolAsset.Weight.ToDec(),
		tokenOut.Amount.ToDec(),
		pool.GetPoolSwapFee(),
	).TruncateInt()
	if tokenInAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenInAmount.GT(tokenInMaxAmount) {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrLimitMaxAmount, "%s token is larger than max amount", outPoolAsset.Token.Denom)
	}

	inPoolAsset.Token.Amount = inPoolAsset.Token.Amount.Add(tokenInAmount)
	outPoolAsset.Token.Amount = outPoolAsset.Token.Amount.Sub(tokenOut.Amount)

	tokenIn := sdk.Coin{Denom: tokenInDenom, Amount: tokenInAmount}

	err = k.updatePoolForSwap(ctx, pool, sender, inPoolAsset, outPoolAsset, tokenIn, tokenOut)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}
	return tokenInAmount, spotPriceAfter, nil
}

// updatePoolForSwap takes a pool, sender, post-swap pool reserves, and tokenIn, tokenOut amounts
// It then updates the pool's balances to the new reserve amounts, and
// sends the in tokens from the sender to the pool, and the out tokens from the pool to the sender.
func (k Keeper) updatePoolForSwap(
	ctx sdk.Context,
	pool types.PoolI,
	sender sdk.AccAddress,
	updatedPoolAssetIn types.PoolAsset,
	updatedPoolAssetOut types.PoolAsset,
	tokenIn sdk.Coin,
	tokenOut sdk.Coin,
) error {
	err := pool.UpdatePoolAssetBalances(sdk.NewCoins(
		updatedPoolAssetIn.Token,
		updatedPoolAssetOut.Token,
	))
	if err != nil {
		return err
	}
	err = k.SetPool(ctx, pool)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), sdk.Coins{
		tokenIn,
	})
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoins(ctx, pool.GetAddress(), sender, sdk.Coins{
		tokenOut,
	})
	if err != nil {
		return err
	}

	tokensIn := sdk.Coins{tokenIn}
	tokensOut := sdk.Coins{tokenOut}
	k.createSwapEvent(ctx, sender, pool.GetId(), tokensIn, tokensOut)
	k.hooks.AfterSwap(ctx, sender, pool.GetId(), tokensIn, tokensOut)
	k.RecordTotalLiquidityIncrease(ctx, tokensIn)
	k.RecordTotalLiquidityDecrease(ctx, tokensOut)

	return err
}

func (k Keeper) CalculateSpotPriceWithSwapFee(ctx sdk.Context, poolId uint64, tokenInDenom, tokenOutDenom string) (sdk.Dec, error) {
	pool, inPoolAsset, outPoolAsset, err :=
		k.getPoolAndInOutAssets(ctx, poolId, tokenInDenom, tokenOutDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	return calcSpotPriceWithSwapFee(
		inPoolAsset.Token.Amount.ToDec(),
		inPoolAsset.Weight.ToDec(),
		outPoolAsset.Token.Amount.ToDec(),
		outPoolAsset.Weight.ToDec(),
		pool.GetPoolSwapFee(),
	), nil
}

func (k Keeper) CalculateSpotPrice(ctx sdk.Context, poolId uint64, tokenInDenom, tokenOutDenom string) (sdk.Dec, error) {
	_, inPoolAsset, outPoolAsset, err :=
		k.getPoolAndInOutAssets(ctx, poolId, tokenInDenom, tokenOutDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	// calcSpotPriceWithSwapFee, but with fee = 0
	return calcSpotPriceWithSwapFee(
		inPoolAsset.Token.Amount.ToDec(),
		inPoolAsset.Weight.ToDec(),
		outPoolAsset.Token.Amount.ToDec(),
		outPoolAsset.Weight.ToDec(),
		sdk.ZeroDec(),
	), nil
}
