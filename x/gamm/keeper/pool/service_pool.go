package pool

import (
	"fmt"

	"github.com/c-osmosis/osmosis/x/gamm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type LiquidityPoolTransactor interface {
	CreatePool(sdk.Context, sdk.AccAddress, sdk.Dec, types.LPTokenInfo, []types.BindTokenInfo) (uint64, error)
	JoinPool(sdk.Context, sdk.AccAddress, uint64, sdk.Int, []types.MaxAmountIn) error
	JoinPoolWithExternAmountIn(sdk.Context, sdk.AccAddress, uint64, string, sdk.Int, sdk.Int) (sdk.Int, error)
	JoinPoolWithPoolAmountOut(sdk.Context, sdk.AccAddress, uint64, string, sdk.Int, sdk.Int) (sdk.Int, error)
	ExitPool(sdk.Context, sdk.AccAddress, uint64, sdk.Int, []types.MinAmountOut) error
	ExitPoolWithPoolAmountIn(sdk.Context, sdk.AccAddress, uint64, string, sdk.Int, sdk.Int) (sdk.Int, error)
	ExitPoolWithExternAmountOut(sdk.Context, sdk.AccAddress, uint64, string, sdk.Int, sdk.Int) (sdk.Int, error)
}

var _ LiquidityPoolTransactor = poolService{}

func (p poolService) CreatePool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	swapFee sdk.Dec,
	lpToken types.LPTokenInfo,
	bindTokens []types.BindTokenInfo,
) (uint64, error) {
	if len(bindTokens) < 2 {
		return 0, sdkerrors.Wrapf(
			types.ErrInvalidRequest,
			"token info length should be at least 2",
		)
	}

	records := make(map[string]types.Record, len(bindTokens))
	for _, info := range bindTokens {
		records[info.Denom] = types.Record{
			DenormalizedWeight: info.Weight,
			Balance:            info.Amount,
		}
	}

	poolId := p.store.GetNextPoolNumber(ctx)
	if lpToken.Denom == "" {
		lpToken.Denom = fmt.Sprintf("osmosis/pool/%d", poolId)
	} else {
		lpToken.Denom = fmt.Sprintf("osmosis/custom/%s", lpToken.Denom)
	}

	pool := types.Pool{
		Id:      poolId,
		SwapFee: swapFee,
		Token: types.LP{
			Denom:       lpToken.Denom,
			Description: lpToken.Description,
			TotalSupply: sdk.NewInt(0),
		},
		TotalWeight: sdk.NewInt(0),
		Records:     records,
	}

	p.store.StorePool(ctx, pool)

	var coins sdk.Coins
	for denom, record := range records {
		coins = append(coins, sdk.Coin{
			Denom:  denom,
			Amount: record.Balance,
		})
	}
	if coins == nil {
		panic("oh my god")
	}
	coins = coins.Sort()

	if err := p.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		sender,
		types.ModuleName,
		coins,
	); err != nil {
		return 0, err
	}

	initialSupply := sdk.NewIntWithDecimal(100, 18)
	lp := lpService{
		denom:      pool.Token.Denom,
		bankKeeper: p.bankKeeper,
	}
	if err := lp.mintPoolShare(ctx, initialSupply); err != nil {
		return 0, err
	}
	if err := lp.pushPoolShare(ctx, sender, initialSupply); err != nil {
		return 0, err
	}
	return pool.Id, nil
}

func (p poolService) joinPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.Pool,
	swapTargets sdk.Coins,
	swapAmount sdk.Int,
) error {
	// process token transfers
	poolShare := lpService{
		denom:      pool.Token.Denom,
		bankKeeper: p.bankKeeper,
	}
	if err := poolShare.mintPoolShare(ctx, swapAmount); err != nil {
		return err
	}
	if err := poolShare.pushPoolShare(ctx, sender, swapAmount); err != nil {
		return err
	}
	if err := p.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		sender,
		types.ModuleName,
		swapTargets,
	); err != nil {
		return err
	}

	// save changes
	pool.Token.TotalSupply = pool.Token.TotalSupply.Add(swapAmount)
	for _, target := range swapTargets {
		record := pool.Records[target.Denom]
		record.Balance = record.Balance.Add(target.Amount)
		pool.Records[target.Denom] = record
	}
	p.store.StorePool(ctx, pool)
	return nil
}

func (p poolService) JoinPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	targetPoolId uint64,
	poolAmountOut sdk.Int,
	maxAmountsIn []types.MaxAmountIn,
) error {
	pool, err := p.store.FetchPool(ctx, targetPoolId)
	if err != nil {
		return err
	}
	lpToken := pool.Token

	poolTotal := lpToken.TotalSupply.ToDec()
	poolRatio := poolAmountOut.ToDec().Quo(poolTotal)
	if poolRatio.Equal(sdk.NewDec(0)) {
		return sdkerrors.Wrapf(types.ErrMathApprox, "calc poolRatio")
	}

	checker := map[string]bool{}
	for _, m := range maxAmountsIn {
		if check := checker[m.Denom]; check {
			return sdkerrors.Wrapf(
				types.ErrInvalidRequest,
				"do not use duplicated denom",
			)
		}
		checker[m.Denom] = true
	}
	if len(pool.Records) != len(checker) {
		return sdkerrors.Wrapf(
			types.ErrInvalidRequest,
			"invalid maxAmountsIn argument",
		)
	}

	var swapTargets sdk.Coins
	for _, maxAmountIn := range maxAmountsIn {
		var (
			tokenDenom    = maxAmountIn.Denom
			record, ok    = pool.Records[tokenDenom]
			tokenAmountIn = poolRatio.Mul(record.Balance.ToDec()).TruncateInt()
		)
		if !ok {
			return sdkerrors.Wrapf(types.ErrInvalidRequest, "token is not bound to pool")
		}
		if tokenAmountIn.Equal(sdk.NewInt(0)) {
			return sdkerrors.Wrapf(types.ErrMathApprox, "calc tokenAmountIn")
		}
		if tokenAmountIn.GT(maxAmountIn.MaxAmount) {
			return sdkerrors.Wrapf(types.ErrLimitExceed, "max amount limited")
		}
		record.Balance = record.Balance.Add(tokenAmountIn)
		pool.Records[tokenDenom] = record // update record

		swapTargets = append(swapTargets, sdk.Coin{
			Denom:  tokenDenom,
			Amount: tokenAmountIn,
		})
	}
	return p.joinPool(ctx, sender, pool, swapTargets, poolAmountOut)
}

func (p poolService) JoinPoolWithExternAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	targetPoolId uint64,
	tokenIn string,
	tokenAmountIn sdk.Int,
	minPoolAmountOut sdk.Int,
) (sdk.Int, error) {
	pool, err := p.store.FetchPool(ctx, targetPoolId)
	if err != nil {
		return sdk.Int{}, err
	}

	record, ok := pool.Records[tokenIn]
	if !ok {
		return sdk.Int{}, sdkerrors.Wrapf(
			types.ErrNotBound,
			"token %s is not bound to this pool", tokenIn,
		)
	}

	// TODO:
	// require(tokenAmountIn <= bmul(_records[tokenIn].balance, MAX_IN_RATIO), "ERR_MAX_IN_RATIO");

	poolAmountOut := calcPoolOutGivenSingleIn(
		record.Balance.ToDec(),
		record.DenormalizedWeight,
		pool.Token.TotalSupply.ToDec(),
		pool.TotalWeight.ToDec(),
		tokenAmountIn.ToDec(),
		pool.SwapFee,
	).TruncateInt()

	if poolAmountOut.LT(minPoolAmountOut) {
		return sdk.Int{}, sdkerrors.Wrapf(
			types.ErrLimitOut,
			"poolShare minimum limit has exceeded",
		)
	}

	if err := p.joinPool(
		ctx,
		sender,
		pool,
		sdk.Coins{{tokenIn, tokenAmountIn}},
		poolAmountOut,
	); err != nil {
		return sdk.Int{}, err
	}

	return poolAmountOut, nil
}

func (p poolService) JoinPoolWithPoolAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	targetPoolId uint64,
	tokenIn string,
	poolAmountOut sdk.Int,
	maxAmountIn sdk.Int,
) (sdk.Int, error) {
	pool, err := p.store.FetchPool(ctx, targetPoolId)
	if err != nil {
		return sdk.Int{}, err
	}

	record, ok := pool.Records[tokenIn]
	if !ok {
		return sdk.Int{}, sdkerrors.Wrapf(
			types.ErrNotBound,
			"token %s is not bound to this pool", tokenIn,
		)
	}

	tokenAmountIn := calcSingleInGivenPoolOut(
		record.Balance.ToDec(),
		record.DenormalizedWeight,
		pool.Token.TotalSupply.ToDec(),
		pool.TotalWeight.ToDec(),
		poolAmountOut.ToDec(),
		pool.SwapFee,
	).TruncateInt()
	if tokenAmountIn.Equal(sdk.NewInt(0)) {
		return sdk.Int{}, sdkerrors.Wrapf(
			types.ErrMathApprox,
			"calculate tokenAmountIn",
		)
	}
	if tokenAmountIn.GT(maxAmountIn) {
		return sdk.Int{}, sdkerrors.Wrapf(
			types.ErrLimitIn,
			"tokenAmount maximum limit has exceeded",
		)
	}

	// TODO:
	// require(tokenAmountIn <= bmul(_records[tokenIn].balance, MAX_IN_RATIO), "ERR_MAX_IN_RATIO");

	if err := p.joinPool(
		ctx,
		sender,
		pool,
		sdk.Coins{{tokenIn, tokenAmountIn}},
		poolAmountOut,
	); err != nil {
		return sdk.Int{}, err
	}

	return poolAmountOut, nil
}

func (p poolService) exitPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.Pool,
	swapTarget sdk.Int,
	swapAmounts sdk.Coins,
) error {
	poolShare := lpService{
		denom:      pool.Token.Denom,
		bankKeeper: p.bankKeeper,
	}
	if err := poolShare.pullPoolShare(ctx, sender, swapTarget); err != nil {
		return err
	}
	if err := poolShare.burnPoolShare(ctx, swapTarget); err != nil {
		return err
	}
	err := p.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		sender,
		swapAmounts,
	)
	if err != nil {
		return err
	}

	// save changes
	pool.Token.TotalSupply = pool.Token.TotalSupply.Sub(swapTarget)
	for _, target := range swapAmounts {
		record := pool.Records[target.Denom]
		record.Balance = record.Balance.Sub(target.Amount)
		pool.Records[target.Denom] = record
	}
	p.store.StorePool(ctx, pool)
	return nil
}

func (p poolService) ExitPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	targetPoolId uint64,
	poolAmountIn sdk.Int,
	minAmountsOut []types.MinAmountOut,
) error {
	pool, err := p.store.FetchPool(ctx, targetPoolId)
	if err != nil {
		return err
	}
	lpToken := pool.Token

	poolTotal := lpToken.TotalSupply.ToDec()
	poolRatio := poolAmountIn.ToDec().Quo(poolTotal)
	if poolRatio.Equal(sdk.NewDec(0)) {
		return sdkerrors.Wrapf(types.ErrMathApprox, "calc poolRatio")
	}

	checker := map[string]bool{}
	for _, m := range minAmountsOut {
		if check := checker[m.Denom]; check {
			return sdkerrors.Wrapf(
				types.ErrInvalidRequest,
				"do not use duplicated denom",
			)
		}
		checker[m.Denom] = true
	}
	if len(pool.Records) != len(checker) {
		return sdkerrors.Wrapf(
			types.ErrInvalidRequest,
			"invalid minAmountsOut argument",
		)
	}

	var swapAmounts sdk.Coins
	for _, minAmountOut := range minAmountsOut {
		var (
			tokenDenom     = minAmountOut.Denom
			record, ok     = pool.Records[tokenDenom]
			tokenAmountOut = poolRatio.Mul(record.Balance.ToDec()).TruncateInt()
		)
		if !ok {
			return sdkerrors.Wrapf(types.ErrInvalidRequest, "token is not bound to pool")
		}
		if tokenAmountOut.Equal(sdk.NewInt(0)) {
			return sdkerrors.Wrapf(types.ErrMathApprox, "calc tokenAmountOut")
		}
		if tokenAmountOut.LT(minAmountOut.MinAmount) {
			return sdkerrors.Wrapf(types.ErrLimitExceed, "min amount limited")
		}
		record.Balance = record.Balance.Sub(tokenAmountOut)
		pool.Records[tokenDenom] = record

		swapAmounts = append(swapAmounts, sdk.Coin{
			Denom:  tokenDenom,
			Amount: tokenAmountOut,
		})
	}
	return p.exitPool(ctx, sender, pool, poolAmountIn, swapAmounts)
}

func (p poolService) ExitPoolWithPoolAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	targetPoolId uint64,
	tokenOut string,
	poolAmountIn sdk.Int,
	minAmountOut sdk.Int,
) (sdk.Int, error) {
	pool, err := p.store.FetchPool(ctx, targetPoolId)
	if err != nil {
		return sdk.Int{}, err
	}

	record, ok := pool.Records[tokenOut]
	if !ok {
		return sdk.Int{}, sdkerrors.Wrapf(
			types.ErrNotBound,
			"token %s is not bound to this pool", tokenOut,
		)
	}

	tokenAmountOut := calcSingleOutGivenPoolIn(
		record.Balance.ToDec(),
		record.DenormalizedWeight,
		pool.Token.TotalSupply.ToDec(),
		pool.TotalWeight.ToDec(),
		poolAmountIn.ToDec(),
		pool.SwapFee,
	).TruncateInt()
	if tokenAmountOut.LT(minAmountOut) {
		return sdk.Int{}, sdkerrors.Wrapf(
			types.ErrLimitOut,
			"tokenAmount minimum limit has exceeded",
		)
	}

	// TODO:
	// require(tokenAmountOut <= bmul(_records[tokenOut].balance, MAX_OUT_RATIO), "ERR_MAX_OUT_RATIO");

	if err := p.exitPool(
		ctx,
		sender,
		pool,
		poolAmountIn,
		sdk.Coins{{tokenOut, tokenAmountOut}},
	); err != nil {
		return sdk.Int{}, err
	}

	return tokenAmountOut, nil
}

func (p poolService) ExitPoolWithExternAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	targetPoolId uint64,
	tokenOut string,
	tokenAmountOut sdk.Int,
	maxPoolAmountIn sdk.Int,
) (sdk.Int, error) {
	pool, err := p.store.FetchPool(ctx, targetPoolId)
	if err != nil {
		return sdk.Int{}, err
	}

	record, ok := pool.Records[tokenOut]
	if !ok {
		return sdk.Int{}, sdkerrors.Wrapf(
			types.ErrNotBound,
			"token %s is not bound to this pool", tokenOut,
		)
	}

	// TOOD:
	// require(tokenAmountOut <= bmul(_records[tokenOut].balance, MAX_OUT_RATIO), "ERR_MAX_OUT_RATIO");

	poolAmountIn := calcPoolInGivenSingleOut(
		record.Balance.ToDec(),
		record.DenormalizedWeight,
		pool.Token.TotalSupply.ToDec(),
		pool.TotalWeight.ToDec(),
		tokenAmountOut.ToDec(),
		pool.SwapFee,
	).TruncateInt()
	if poolAmountIn.Equal(sdk.NewInt(0)) {
		return sdk.Int{}, sdkerrors.Wrapf(
			types.ErrMathApprox,
			"calculate poolAmountIn",
		)
	}
	if poolAmountIn.GT(maxPoolAmountIn) {
		return sdk.Int{}, sdkerrors.Wrapf(
			types.ErrLimitIn,
			"poolAmount maximum limit has exceeded",
		)
	}

	if err := p.exitPool(
		ctx,
		sender,
		pool,
		poolAmountIn,
		sdk.Coins{{tokenOut, tokenAmountOut}},
	); err != nil {
		return sdk.Int{}, err
	}

	return poolAmountIn, nil
}
