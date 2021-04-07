package keeper

import (
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
	poolAcc, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	if poolAcc.GetPoolParams().Lock {
		return sdk.Int{}, sdk.Dec{}, types.ErrPoolLocked
	}

	inRecord, err := poolAcc.GetRecord(tokenIn.Denom)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	outRecord, err := poolAcc.GetRecord(tokenOutDenom)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	// TODO: Understand if we are handling swap fee consistently, with the global swap fee and the pool swap fee
	//
	spotPriceBefore := calcSpotPriceWithSwapFee(
		inRecord.Token.Amount.ToDec(),
		inRecord.Weight.ToDec(),
		outRecord.Token.Amount.ToDec(),
		outRecord.Weight.ToDec(),
		poolAcc.GetPoolParams().SwapFee,
	)

	tokenOutAmount = calcOutGivenIn(
		inRecord.Token.Amount.ToDec(),
		inRecord.Weight.ToDec(),
		outRecord.Token.Amount.ToDec(),
		outRecord.Weight.ToDec(),
		tokenIn.Amount.ToDec(),
		poolAcc.GetPoolParams().SwapFee,
	).TruncateInt()
	if tokenOutAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenOutAmount.LT(tokenOutMinAmount) {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrLimitMinAmount, "%s token is lesser than min amount", outRecord.Token.Denom)
	}

	inRecord.Token.Amount = inRecord.Token.Amount.Add(tokenIn.Amount)
	outRecord.Token.Amount = outRecord.Token.Amount.Sub(tokenOutAmount)

	spotPriceAfter = calcSpotPriceWithSwapFee(
		inRecord.Token.Amount.ToDec(),
		inRecord.Weight.ToDec(),
		outRecord.Token.Amount.ToDec(),
		outRecord.Weight.ToDec(),
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

	err = poolAcc.SetRecords([]types.Record{
		inRecord,
		outRecord,
	})
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
	poolAcc, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	if poolAcc.GetPoolParams().Lock {
		return sdk.Int{}, sdk.Dec{}, types.ErrPoolLocked
	}

	inRecord, err := poolAcc.GetRecord(tokenInDenom)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	outRecord, err := poolAcc.GetRecord(tokenOut.Denom)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	spotPriceBefore := calcSpotPriceWithSwapFee(
		inRecord.Token.Amount.ToDec(),
		inRecord.Weight.ToDec(),
		outRecord.Token.Amount.ToDec(),
		outRecord.Weight.ToDec(),
		poolAcc.GetPoolParams().SwapFee,
	)

	tokenInAmount = calcInGivenOut(
		inRecord.Token.Amount.ToDec(),
		inRecord.Weight.ToDec(),
		outRecord.Token.Amount.ToDec(),
		outRecord.Weight.ToDec(),
		tokenOut.Amount.ToDec(),
		poolAcc.GetPoolParams().SwapFee,
	).TruncateInt()
	if tokenInAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenInAmount.GT(tokenInMaxAmount) {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrLimitMaxAmount, "%s token is larger than max amount", outRecord.Token.Denom)
	}

	inRecord.Token.Amount = inRecord.Token.Amount.Add(tokenInAmount)
	outRecord.Token.Amount = outRecord.Token.Amount.Sub(tokenOut.Amount)

	spotPriceAfter = calcSpotPriceWithSwapFee(
		inRecord.Token.Amount.ToDec(),
		inRecord.Weight.ToDec(),
		outRecord.Token.Amount.ToDec(),
		outRecord.Weight.ToDec(),
		poolAcc.GetPoolParams().SwapFee,
	)
	if spotPriceAfter.LT(spotPriceBefore) {
		return sdk.Int{}, sdk.Dec{}, types.ErrInvalidMathApprox
	}
	if spotPriceBefore.GT(tokenInAmount.ToDec().QuoInt(tokenOut.Amount)) {
		return sdk.Int{}, sdk.Dec{}, types.ErrInvalidMathApprox
	}

	err = poolAcc.SetRecords([]types.Record{
		inRecord,
		outRecord,
	})
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

	return tokenInAmount, spotPriceAfter, nil
}

func (k Keeper) CalculateSpotPrice(ctx sdk.Context, poolId uint64, tokenInDenom, tokenOutDenom string) (sdk.Dec, error) {
	poolAcc, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	inRecord, err := poolAcc.GetRecord(tokenInDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	outRecord, err := poolAcc.GetRecord(tokenOutDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	return calcSpotPriceWithSwapFee(
		inRecord.Token.Amount.ToDec(),
		inRecord.Weight.ToDec(),
		outRecord.Token.Amount.ToDec(),
		outRecord.Weight.ToDec(),
		poolAcc.GetPoolParams().SwapFee,
	), nil
}

func (k Keeper) CalculateSpotPriceSansSwapFee(ctx sdk.Context, poolId uint64, tokenInDenom, tokenOutDenom string) (sdk.Dec, error) {
	poolAcc, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	inRecord, err := poolAcc.GetRecord(tokenInDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	outRecord, err := poolAcc.GetRecord(tokenOutDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	return calcSpotPriceWithSwapFee(
		inRecord.Token.Amount.ToDec(),
		inRecord.Weight.ToDec(),
		outRecord.Token.Amount.ToDec(),
		outRecord.Weight.ToDec(),
		sdk.ZeroDec(),
	), nil
}
