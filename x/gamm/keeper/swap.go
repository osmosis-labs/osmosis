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
) (tokenOutAmount sdk.Int, err error) {
	if tokenIn.Denom == tokenOutDenom {
		return sdk.Int{}, errors.New("cannot trade same denomination in and out")
	}
	tokensIn := sdk.Coins{tokenIn}

	pool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}
	swapFee := pool.GetSwapFee(ctx)

	if !pool.IsActive(ctx.BlockTime()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrPoolLocked, "swap on inactive pool")
	}

	tokenOutDecCoin, err := pool.CalcOutAmtGivenIn(ctx, tokensIn, tokenOutDenom, swapFee)
	if err != nil {
		return sdk.Int{}, err
	}
	tokenOutCoin, _ := tokenOutDecCoin.TruncateDecimal()
	tokenOutAmount = tokenOutCoin.Amount
	if tokenOutAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenOutAmount.LT(tokenOutMinAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMinAmount, "%s token is lesser than min amount", tokenOutDenom)
	}

	err = k.updatePoolForSwap(ctx, pool, sender, tokenIn, tokenOutCoin)
	if err != nil {
		return sdk.Int{}, err
	}

	return tokenOutAmount, nil
}

func (k Keeper) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	if tokenInDenom == tokenOut.Denom {
		return sdk.Int{}, errors.New("cannot trade same denomination in and out")
	}

	pool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}
	swapFee := pool.GetSwapFee(ctx)

	if !pool.IsActive(ctx.BlockTime()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrPoolLocked, "swap on inactive pool")
	}

	poolOutBal := pool.GetTotalLpBalances(ctx).AmountOf(tokenOut.Denom)
	if tokenOut.Amount.GTE(poolOutBal) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrTooManyTokensOut,
			"can't get more tokens out than there are tokens in the pool")
	}

	tokenInDecCoin, err := pool.CalcInAmtGivenOut(ctx, sdk.Coins{tokenOut}, tokenInDenom, swapFee)
	if err != nil {
		return sdk.Int{}, err
	}
	tokenInCoin, _ := tokenInDecCoin.TruncateDecimal()
	tokenInAmount = tokenInCoin.Amount

	if tokenInAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenInAmount.GT(tokenInMaxAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMaxAmount, "%s token required is larger than max amount", tokenInDenom)
	}

	tokenIn := sdk.Coin{Denom: tokenInDenom, Amount: tokenInAmount}

	err = k.updatePoolForSwap(ctx, pool, sender, tokenIn, tokenOut)
	if err != nil {
		return sdk.Int{}, err
	}
	return tokenInAmount, nil
}

// updatePoolForSwap takes a pool, sender, and tokenIn, tokenOut amounts
// It then updates the pool's balances to the new reserve amounts, and
// sends the in tokens from the sender to the pool, and the out tokens from the pool to the sender.
func (k Keeper) updatePoolForSwap(
	ctx sdk.Context,
	pool types.PoolI,
	sender sdk.AccAddress,
	tokenIn sdk.Coin,
	tokenOut sdk.Coin,
) error {
	tokensIn := sdk.Coins{tokenIn}
	tokensOut := sdk.Coins{tokenOut}

	err := pool.ApplySwap(ctx, tokensIn, tokensOut)
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

	k.createSwapEvent(ctx, sender, pool.GetId(), tokensIn, tokensOut)
	k.hooks.AfterSwap(ctx, sender, pool.GetId(), tokensIn, tokensOut)
	k.RecordTotalLiquidityIncrease(ctx, tokensIn)
	k.RecordTotalLiquidityDecrease(ctx, tokensOut)

	return err
}
