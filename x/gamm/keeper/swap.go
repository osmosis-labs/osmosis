package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/c-osmosis/osmosis/x/gamm/types"
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

	poolAcc, inPoolAsset, outPoolAsset, err :=
		k.getPoolAndInOutAssets(ctx, poolId, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	// TODO: Understand if we are handling swap fee consistently,
	// with the global swap fee and the pool swap fee

	tokenOutAmount = calcOutGivenIn(
		inPoolAsset.Token.Amount.ToDec(),
		inPoolAsset.Weight.ToDec(),
		outPoolAsset.Token.Amount.ToDec(),
		outPoolAsset.Weight.ToDec(),
		tokenIn.Amount.ToDec(),
		poolAcc.GetPoolParams().SwapFee,
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

	err = k.updatePoolForSwap(ctx, poolAcc, sender, inPoolAsset, outPoolAsset, tokenIn, tokenOut)
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

	poolAcc, inPoolAsset, outPoolAsset, err :=
		k.getPoolAndInOutAssets(ctx, poolId, tokenInDenom, tokenOut.Denom)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	tokenInAmount = calcInGivenOut(
		inPoolAsset.Token.Amount.ToDec(),
		inPoolAsset.Weight.ToDec(),
		outPoolAsset.Token.Amount.ToDec(),
		outPoolAsset.Weight.ToDec(),
		tokenOut.Amount.ToDec(),
		poolAcc.GetPoolParams().SwapFee,
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

	err = k.updatePoolForSwap(ctx, poolAcc, sender, inPoolAsset, outPoolAsset, tokenIn, tokenOut)
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
	poolAcc types.PoolI,
	sender sdk.AccAddress,
	updatedPoolAssetIn types.PoolAsset,
	updatedPoolAssetOut types.PoolAsset,
	tokenIn sdk.Coin,
	tokenOut sdk.Coin,
) error {
	err := poolAcc.UpdatePoolAssetBalances(sdk.NewCoins(
		updatedPoolAssetIn.Token,
		updatedPoolAssetOut.Token,
	))
	if err != nil {
		return err
	}
	err = k.SetPool(ctx, poolAcc)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoins(ctx, sender, poolAcc.GetAddress(), sdk.Coins{
		tokenIn,
	})
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoins(ctx, poolAcc.GetAddress(), sender, sdk.Coins{
		tokenOut,
	})
	if err != nil {
		return err
	}

	k.hooks.AfterSwap(ctx, sender, poolAcc.GetId(), sdk.Coins{tokenIn}, sdk.Coins{tokenOut})

	return err
}

func (k Keeper) CalculateSpotPrice(ctx sdk.Context, poolId uint64, tokenInDenom, tokenOutDenom string) (sdk.Dec, error) {
	poolAcc, inPoolAsset, outPoolAsset, err :=
		k.getPoolAndInOutAssets(ctx, poolId, tokenInDenom, tokenOutDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	return calcSpotPriceWithSwapFee(
		inPoolAsset.Token.Amount.ToDec(),
		inPoolAsset.Weight.ToDec(),
		outPoolAsset.Token.Amount.ToDec(),
		outPoolAsset.Weight.ToDec(),
		poolAcc.GetPoolParams().SwapFee,
	), nil
}

func (k Keeper) CalculateSpotPriceSansSwapFee(ctx sdk.Context, poolId uint64, tokenInDenom, tokenOutDenom string) (sdk.Dec, error) {
	_, inPoolAsset, outPoolAsset, err :=
		k.getPoolAndInOutAssets(ctx, poolId, tokenInDenom, tokenOutDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	return calcSpotPriceWithSwapFee(
		inPoolAsset.Token.Amount.ToDec(),
		inPoolAsset.Weight.ToDec(),
		outPoolAsset.Token.Amount.ToDec(),
		outPoolAsset.Weight.ToDec(),
		sdk.ZeroDec(),
	), nil
}
