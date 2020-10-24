package pool

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type Service interface {
	// Viewer
	GetPool(sdk.Context, uint64) (types.Pool, error)
	GetSwapFee(sdk.Context, uint64) (sdk.Dec, error)
	GetShareInfo(sdk.Context, uint64) (types.LP, error)
	GetTokenBalance(sdk.Context, uint64) (sdk.Coins, error)
	GetSpotPrice(sdk.Context, uint64, string, string) (sdk.Int, error)

	// Sender
	LiquidityPoolTransactor
	SwapExactAmountIn(sdk.Context, sdk.AccAddress, uint64, sdk.Coin, sdk.Int, sdk.Coin, sdk.Int, sdk.Int) (sdk.Dec, sdk.Dec, error)
	SwapExactAmountOut(sdk.Context, sdk.AccAddress, uint64, sdk.Coin, sdk.Int, sdk.Coin, sdk.Int, sdk.Int) (sdk.Dec, sdk.Dec, error)
}

type poolService struct {
	store         Store
	accountKeeper types.AccountKeeper
	bankKeeper    bankkeeper.Keeper
}

func NewService(
	store Store,
	accountKeeper types.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
) Service {
	return poolService{
		store:         store,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (p poolService) GetPool(ctx sdk.Context, poolId uint64) (types.Pool, error) {
	pool, err := p.store.FetchPool(ctx, poolId)
	if err != nil {
		return types.Pool{}, err
	}
	return pool, nil
}

func (p poolService) GetSwapFee(ctx sdk.Context, poolId uint64) (sdk.Dec, error) {
	pool, err := p.store.FetchPool(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}
	return pool.SwapFee, nil
}

func (p poolService) GetShareInfo(ctx sdk.Context, poolId uint64) (types.LP, error) {
	pool, err := p.store.FetchPool(ctx, poolId)
	if err != nil {
		return types.LP{}, err
	}
	return pool.Token, nil
}

func (p poolService) GetTokenBalance(ctx sdk.Context, poolId uint64) (sdk.Coins, error) {
	pool, err := p.store.FetchPool(ctx, poolId)
	if err != nil {
		return nil, err
	}

	var coins sdk.Coins
	for denom, record := range pool.Records {
		coins = append(coins, sdk.Coin{
			Denom:  denom,
			Amount: record.Balance,
		})
	}
	if coins == nil {
		panic("oh my god")
	}
	coins = coins.Sort()

	return coins, nil
}

func (p poolService) GetSpotPrice(ctx sdk.Context, poolId uint64, tokenIn, tokenOut string) (sdk.Int, error) {
	pool, err := p.store.FetchPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	inRecord, ok := pool.Records[tokenIn]
	if !ok {
		return sdk.Int{}, sdkerrors.Wrapf(
			types.ErrNotBound,
			"token %s is not bound to pool", tokenIn,
		)
	}
	outRecord, ok := pool.Records[tokenOut]
	if !ok {
		return sdk.Int{}, sdkerrors.Wrapf(
			types.ErrNotBound,
			"token %s is not bound to pool", tokenOut,
		)
	}

	spotPrice := calcSpotPrice(
		inRecord.Balance.ToDec(),
		inRecord.DenormalizedWeight,
		outRecord.Balance.ToDec(),
		outRecord.DenormalizedWeight,
		pool.SwapFee,
	).TruncateInt()

	return spotPrice, nil
}

func (p poolService) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	targetPoolId uint64,
	tokenIn sdk.Coin,
	tokenAmountIn sdk.Int,
	tokenOut sdk.Coin,
	minAmountOut sdk.Int,
	maxPrice sdk.Int) (tokenAmountOut sdk.Dec, spotPriceAfter sdk.Dec, err error) {

	pool, err := p.store.FetchPool(ctx, targetPoolId)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	inRecord := pool.Records[tokenIn.Denom]
	outRecord := pool.Records[tokenOut.Denom]

	// todo: require(tokenAmountIn <= bmul(inRecord.balance, MAX_IN_RATIO), "ERR_MAX_IN_RATIO");
	if true /* todo: bmul(inRecord.balance, MAX_IN_RATIO) sdk.Dec.GTE(tokenAmountIn.ToDec()) */ {
		return sdk.Dec{}, sdk.Dec{}, types.ErrMaxInRatio
	}

	// 1.
	spotPriceBefore := calcSpotPrice(
		inRecord.Balance.ToDec(),
		inRecord.DenormalizedWeight,
		outRecord.Balance.ToDec(),
		outRecord.DenormalizedWeight,
		pool.SwapFee,
	)
	if maxPrice.ToDec().GTE(spotPriceBefore) {
		return sdk.Dec{}, sdk.Dec{}, types.ErrBadLimitPrice
	}

	// 2.
	tokenAmountOut = calcOutGivenIn(
		inRecord.Balance.ToDec(),
		inRecord.DenormalizedWeight,
		outRecord.Balance.ToDec(),
		outRecord.DenormalizedWeight,
		tokenAmountIn.ToDec(),
		pool.SwapFee,
	)
	if tokenAmountOut.GTE(minAmountOut.ToDec()) {
		return sdk.Dec{}, sdk.Dec{}, types.ErrLimitOut
	}

	//todo: inRecord.balance = badd(inRecord.balance, tokenAmountIn);
	//todo: outRecord.balance = bsub(outRecord.balance, tokenAmountOut);

	// 3.
	spotPriceAfter = calcSpotPrice(
		inRecord.Balance.ToDec(),
		inRecord.DenormalizedWeight,
		outRecord.Balance.ToDec(),
		outRecord.DenormalizedWeight,
		pool.SwapFee,
	)
	if spotPriceAfter.GTE(spotPriceBefore) {
		return sdk.Dec{}, sdk.Dec{}, types.ErrMathApprox
	}
	if maxPrice.ToDec().GTE(spotPriceAfter) {
		return sdk.Dec{}, sdk.Dec{}, types.ErrLimitPrice
	}
	if /* todo: bdiv(tokenAmountIn, tokenAmountOut).GTE(spotPriceBefore) */ true {
		return sdk.Dec{}, sdk.Dec{}, types.ErrMathApprox
	}

	inRecord.Balance = inRecord.Balance.Add(tokenAmountIn)
	pool.Records[tokenIn.Denom] = inRecord

	outRecord.Balance = outRecord.Balance.Sub(sdk.Int(tokenAmountOut))
	pool.Records[tokenOut.Denom] = outRecord

	p.store.StorePool(ctx, pool)

	return tokenAmountOut, spotPriceAfter, nil
}

func (p poolService) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	targetPoolId uint64,
	tokenIn sdk.Coin,
	maxAmountIn sdk.Int,
	tokenOut sdk.Coin,
	tokenAmountOut sdk.Int,
	maxPrice sdk.Int) (tokenAmountIn sdk.Dec, spotPriceAfter sdk.Dec, err error) {

	pool, err := p.store.FetchPool(ctx, targetPoolId)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	inRecord := pool.Records[tokenIn.Denom]
	outRecord := pool.Records[tokenOut.Denom]

	if true /*sdk.Dec{}.GTE(tokenAmountOut.ToDec())*/ {
		return sdk.Dec{}, sdk.Dec{}, types.ErrMaxOutRatio
	}

	// 1.
	spotPriceBefore := calcSpotPrice(
		inRecord.Balance.ToDec(),
		inRecord.DenormalizedWeight,
		outRecord.Balance.ToDec(),
		outRecord.DenormalizedWeight,
		pool.SwapFee,
	)
	if maxPrice.ToDec().GTE(spotPriceBefore) {
		return sdk.Dec{}, sdk.Dec{}, types.ErrBadLimitPrice
	}

	// 2.
	tokenAmountIn = calcInGivenOut(
		inRecord.Balance.ToDec(),
		inRecord.DenormalizedWeight,
		outRecord.Balance.ToDec(),
		outRecord.DenormalizedWeight,
		tokenAmountOut.ToDec(),
		pool.SwapFee,
	)
	if maxAmountIn.ToDec().GTE(tokenAmountIn) {
		return sdk.Dec{}, sdk.Dec{}, types.ErrLimitIn
	}

	// todo: inRecord.balance = badd(inRecord.balance, tokenAmountIn);
	// todo: outRecord.balance = bsub(outRecord.balance, tokenAmountOut);

	// 3.
	spotPriceAfter = calcSpotPrice(
		inRecord.Balance.ToDec(),
		inRecord.DenormalizedWeight,
		outRecord.Balance.ToDec(),
		outRecord.DenormalizedWeight,
		pool.SwapFee,
	)
	if spotPriceAfter.GTE(spotPriceBefore) {
		return sdk.Dec{}, sdk.Dec{}, types.ErrMathApprox
	}
	if maxPrice.ToDec().GTE(spotPriceAfter) {
		return sdk.Dec{}, sdk.Dec{}, types.ErrLimitPrice
	}
	if true /* todo: bdiv(tokenAmountIn, tokenAmountOut).GTE(spotPriceBefore) */ {
		return sdk.Dec{}, sdk.Dec{}, types.ErrMathApprox
	}

	inRecord.Balance = inRecord.Balance.Add(sdk.Int(tokenAmountIn))
	pool.Records[tokenIn.Denom] = inRecord

	outRecord.Balance = outRecord.Balance.Sub(tokenAmountOut)
	pool.Records[tokenOut.Denom] = outRecord

	p.store.StorePool(ctx, pool)

	return tokenAmountIn, spotPriceAfter, nil
}
