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

	poolAcc, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	if poolAcc.GetPoolParams().Lock {
		return sdk.Int{}, sdk.Dec{}, types.ErrPoolLocked
	}

	inPoolAsset, err := poolAcc.GetPoolAsset(tokenIn.Denom)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	outPoolAsset, err := poolAcc.GetPoolAsset(tokenOutDenom)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	// TODO: Understand if we are handling swap fee consistently, with the global swap fee and the pool swap fee
	//
	spotPriceBefore := calcSpotPriceWithSwapFee(
		inPoolAsset.Token.Amount.ToDec(),
		inPoolAsset.Weight.ToDec(),
		outPoolAsset.Token.Amount.ToDec(),
		outPoolAsset.Weight.ToDec(),
		poolAcc.GetPoolParams().SwapFee,
	)

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

	spotPriceAfter = calcSpotPriceWithSwapFee(
		inPoolAsset.Token.Amount.ToDec(),
		inPoolAsset.Weight.ToDec(),
		outPoolAsset.Token.Amount.ToDec(),
		outPoolAsset.Weight.ToDec(),
		poolAcc.GetPoolParams().SwapFee,
	)
	if spotPriceAfter.LT(spotPriceBefore) {
		return sdk.Int{}, sdk.Dec{}, types.ErrInvalidMathApprox
	}

	// TODO: Do we need this check, seems pretty expensive?
	// I'd rather spend that computation in ensuring a better approx
	if spotPriceBefore.GT(tokenIn.Amount.ToDec().QuoInt(tokenOutAmount)) {
		return sdk.Int{}, sdk.Dec{}, types.ErrInvalidMathApprox
	}

	err = poolAcc.UpdatePoolAssetBalances(sdk.NewCoins(
		inPoolAsset.Token,
		outPoolAsset.Token,
	))
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}
	err = k.SetPool(ctx, poolAcc)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	err = k.bankKeeper.SendCoins(ctx, sender, poolAcc.GetAddress(), sdk.Coins{tokenIn})
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	err = k.bankKeeper.SendCoins(ctx, poolAcc.GetAddress(), sender, sdk.Coins{
		sdk.NewCoin(tokenOutDenom, tokenOutAmount),
	})
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	k.hooks.AfterSwap(ctx, sender, poolAcc.GetId(), sdk.Coins{tokenIn}, sdk.Coins{sdk.NewCoin(tokenOutDenom, tokenOutAmount)})

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

	poolAcc, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	if poolAcc.GetPoolParams().Lock {
		return sdk.Int{}, sdk.Dec{}, types.ErrPoolLocked
	}

	inPoolAsset, err := poolAcc.GetPoolAsset(tokenInDenom)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	outPoolAsset, err := poolAcc.GetPoolAsset(tokenOut.Denom)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	spotPriceBefore := calcSpotPriceWithSwapFee(
		inPoolAsset.Token.Amount.ToDec(),
		inPoolAsset.Weight.ToDec(),
		outPoolAsset.Token.Amount.ToDec(),
		outPoolAsset.Weight.ToDec(),
		poolAcc.GetPoolParams().SwapFee,
	)

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

	spotPriceAfter = calcSpotPriceWithSwapFee(
		inPoolAsset.Token.Amount.ToDec(),
		inPoolAsset.Weight.ToDec(),
		outPoolAsset.Token.Amount.ToDec(),
		outPoolAsset.Weight.ToDec(),
		poolAcc.GetPoolParams().SwapFee,
	)
	if spotPriceAfter.LT(spotPriceBefore) {
		return sdk.Int{}, sdk.Dec{}, types.ErrInvalidMathApprox
	}
	if spotPriceBefore.GT(tokenInAmount.ToDec().QuoInt(tokenOut.Amount)) {
		return sdk.Int{}, sdk.Dec{}, types.ErrInvalidMathApprox
	}

	err = poolAcc.UpdatePoolAssetBalances(sdk.NewCoins(
		inPoolAsset.Token,
		outPoolAsset.Token,
	))
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}
	err = k.SetPool(ctx, poolAcc)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	err = k.bankKeeper.SendCoins(ctx, sender, poolAcc.GetAddress(), sdk.Coins{
		sdk.NewCoin(tokenInDenom, tokenInAmount),
	})
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	err = k.bankKeeper.SendCoins(ctx, poolAcc.GetAddress(), sender, sdk.Coins{
		tokenOut,
	})
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	k.hooks.AfterSwap(ctx, sender, poolAcc.GetId(), sdk.Coins{sdk.NewCoin(tokenInDenom, tokenInAmount)}, sdk.Coins{tokenOut})

	return tokenInAmount, spotPriceAfter, nil
}

func (k Keeper) CalculateSpotPrice(ctx sdk.Context, poolId uint64, tokenInDenom, tokenOutDenom string) (sdk.Dec, error) {
	poolAcc, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	inPoolAsset, err := poolAcc.GetPoolAsset(tokenInDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	outPoolAsset, err := poolAcc.GetPoolAsset(tokenOutDenom)
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
	poolAcc, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	inPoolAsset, err := poolAcc.GetPoolAsset(tokenInDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	outPoolAsset, err := poolAcc.GetPoolAsset(tokenOutDenom)
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
