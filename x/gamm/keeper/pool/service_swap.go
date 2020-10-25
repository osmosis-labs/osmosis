package pool

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type LiquiditySwapTransactor interface {
	SwapExactAmountIn(sdk.Context, sdk.AccAddress, uint64, sdk.Coin, string, sdk.Int, sdk.Int) (sdk.Int, sdk.Dec, error)
	SwapExactAmountOut(sdk.Context, sdk.AccAddress, uint64, string, sdk.Int, sdk.Coin, sdk.Int) (sdk.Int, sdk.Dec, error)
}

var _ LiquiditySwapTransactor = poolService{}

func (p poolService) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	targetPoolId uint64,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	minAmountOut sdk.Int,
	maxPrice sdk.Int) (tokenAmountOutInt sdk.Int, spotPriceAfter sdk.Dec, err error) {

	pool, err := p.store.FetchPool(ctx, targetPoolId)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}
	inRecord, ok := pool.Records[tokenIn.Denom]
	if !ok {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrDenomNotExist, "%s doesn't exist on pool", tokenIn.Denom)
	}
	outRecord, ok := pool.Records[tokenOutDenom]
	if !ok {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrDenomNotExist, "%s doesn't exist on pool", tokenOutDenom)
	}

	tokenAmountIn := tokenIn.Amount

	if tokenAmountIn.GT(maxInRatio.MulInt(inRecord.Balance).TruncateInt()) {
		return sdk.Int{}, sdk.Dec{}, types.ErrMaxInRatio
	}

	// 1.
	spotPriceBefore := calcSpotPrice(
		inRecord.Balance.ToDec(),
		inRecord.DenormalizedWeight,
		outRecord.Balance.ToDec(),
		outRecord.DenormalizedWeight,
		pool.SwapFee,
	)
	if spotPriceBefore.TruncateInt().GT(maxPrice) {
		return sdk.Int{}, sdk.Dec{}, types.ErrBadLimitPrice
	}

	// 2.
	tokenAmountOut := calcOutGivenIn(
		inRecord.Balance.ToDec(),
		inRecord.DenormalizedWeight,
		outRecord.Balance.ToDec(),
		outRecord.DenormalizedWeight,
		tokenAmountIn.ToDec(),
		pool.SwapFee,
	)
	if tokenAmountOut.TruncateInt().LT(minAmountOut) {
		return sdk.Int{}, sdk.Dec{}, types.ErrLimitOut
	}

	inRecord.Balance = inRecord.Balance.Add(tokenAmountIn)
	outRecord.Balance = outRecord.Balance.Sub(tokenAmountOut.TruncateInt())

	// 3.
	spotPriceAfter = calcSpotPrice(
		inRecord.Balance.ToDec(),
		inRecord.DenormalizedWeight,
		outRecord.Balance.ToDec(),
		outRecord.DenormalizedWeight,
		pool.SwapFee,
	)
	if spotPriceAfter.LT(spotPriceBefore) {
		return sdk.Int{}, sdk.Dec{}, types.ErrMathApprox
	}
	if maxPrice.ToDec().LT(spotPriceAfter) {
		return sdk.Int{}, sdk.Dec{}, types.ErrLimitPrice
	}
	if spotPriceBefore.GT(tokenAmountIn.ToDec().Quo(tokenAmountOut)) {
		return sdk.Int{}, sdk.Dec{}, types.ErrMathApprox
	}

	pool.Records[tokenIn.Denom] = inRecord
	pool.Records[tokenOutDenom] = outRecord

	p.store.StorePool(ctx, pool)

	err = p.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.Coins{
		tokenIn,
	})
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}
	err = p.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, sdk.Coins{
		sdk.NewCoin(tokenOutDenom, tokenAmountOut.TruncateInt()),
	})
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	return tokenAmountOut.TruncateInt(), spotPriceAfter, nil
}

func (p poolService) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	targetPoolId uint64,
	tokenInDenom string,
	maxAmountIn sdk.Int,
	tokenOut sdk.Coin,
	maxPrice sdk.Int) (tokenAmountInInt sdk.Int, spotPriceAfter sdk.Dec, err error) {

	pool, err := p.store.FetchPool(ctx, targetPoolId)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}
	inRecord, ok := pool.Records[tokenInDenom]
	if !ok {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrDenomNotExist, "%s doesn't exist on pool", tokenInDenom)
	}
	outRecord, ok := pool.Records[tokenOut.Denom]
	if !ok {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrDenomNotExist, "%s doesn't exist on pool", tokenOut.Denom)
	}

	if tokenOut.Amount.GT(maxOutRatio.MulInt(outRecord.Balance).TruncateInt()) {
		return sdk.Int{}, sdk.Dec{}, types.ErrMaxOutRatio
	}

	// 1.
	spotPriceBefore := calcSpotPrice(
		inRecord.Balance.ToDec(),
		inRecord.DenormalizedWeight,
		outRecord.Balance.ToDec(),
		outRecord.DenormalizedWeight,
		pool.SwapFee,
	)
	if spotPriceBefore.GT(maxPrice.ToDec()) {
		return sdk.Int{}, sdk.Dec{}, types.ErrBadLimitPrice
	}

	// 2.
	tokenAmountIn := calcInGivenOut(
		inRecord.Balance.ToDec(),
		inRecord.DenormalizedWeight,
		outRecord.Balance.ToDec(),
		outRecord.DenormalizedWeight,
		tokenOut.Amount.ToDec(),
		pool.SwapFee,
	)
	if tokenAmountIn.GT(maxAmountIn.ToDec()) {
		return sdk.Int{}, sdk.Dec{}, types.ErrLimitIn
	}

	inRecord.Balance = inRecord.Balance.Add(tokenAmountIn.TruncateInt())
	outRecord.Balance = outRecord.Balance.Sub(tokenOut.Amount)

	// 3.
	spotPriceAfter = calcSpotPrice(
		inRecord.Balance.ToDec(),
		inRecord.DenormalizedWeight,
		outRecord.Balance.ToDec(),
		outRecord.DenormalizedWeight,
		pool.SwapFee,
	)
	if spotPriceAfter.LT(spotPriceBefore) {
		return sdk.Int{}, sdk.Dec{}, types.ErrMathApprox
	}
	if spotPriceAfter.GT(maxPrice.ToDec()) {
		return sdk.Int{}, sdk.Dec{}, types.ErrLimitPrice
	}
	if spotPriceBefore.GT(tokenAmountIn.QuoInt(tokenOut.Amount)) {
		return sdk.Int{}, sdk.Dec{}, types.ErrMathApprox
	}

	pool.Records[tokenInDenom] = inRecord
	pool.Records[tokenOut.Denom] = outRecord

	p.store.StorePool(ctx, pool)

	err = p.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.Coins{
		sdk.NewCoin(tokenInDenom, tokenAmountIn.TruncateInt()),
	})
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}
	err = p.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, sdk.Coins{
		tokenOut,
	})
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}

	return tokenAmountIn.TruncateInt(), spotPriceAfter, nil
}
