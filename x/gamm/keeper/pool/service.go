package pool

import (
	"fmt"

	"github.com/c-osmosis/osmosis/x/gamm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type Service interface {
	// Viewer
	GetPoolShareInfo(sdk.Context, uint64) (types.LP, error)
	GetPoolTokenBalance(sdk.Context, uint64) (sdk.Coins, error)
	GetSpotPrice(sdk.Context, uint64, string, string) (sdk.Int, error)

	// Sender
	CreatePool(sdk.Context, sdk.AccAddress, sdk.Dec, types.LPTokenInfo, []types.BindTokenInfo) error
	JoinPool(sdk.Context, sdk.AccAddress, uint64, sdk.Int, []types.MaxAmountIn) error
	ExitPool(sdk.Context, sdk.AccAddress, uint64, sdk.Int, []types.MinAmountOut) error
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

func (p poolService) GetPoolShareInfo(ctx sdk.Context, poolId uint64) (types.LP, error) {
	pool, err := p.store.FetchPool(ctx, poolId)
	if err != nil {
		return types.LP{}, err
	}
	return pool.Token, nil
}

func (p poolService) GetPoolTokenBalance(ctx sdk.Context, poolId uint64) (sdk.Coins, error) {
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

func (p poolService) CreatePool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	swapFee sdk.Dec,
	lpToken types.LPTokenInfo,
	bindTokens []types.BindTokenInfo,
) error {
	if len(bindTokens) < 2 {
		return sdkerrors.Wrapf(
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

	return p.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		sender,
		types.ModuleName,
		coins,
	)
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

	var sendTargets sdk.Coins
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

		sendTargets = append(sendTargets, sdk.Coin{
			Denom:  tokenDenom,
			Amount: tokenAmountIn,
		})
	}

	// process token transfer
	err = p.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		sender,
		types.ModuleName,
		sendTargets,
	)
	if err != nil {
		return err
	}

	// process lpToken transfer
	poolShare := lpService{
		denom:      lpToken.Denom,
		bankKeeper: p.bankKeeper,
	}
	if err := poolShare.mintPoolShare(ctx, poolAmountOut); err != nil {
		return err
	}
	if err := poolShare.pushPoolShare(ctx, sender, poolAmountOut); err != nil {
		return err
	}

	// save changes
	lpToken.TotalSupply = lpToken.TotalSupply.Add(poolAmountOut)
	pool.Token = lpToken
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

	var sendTargets sdk.Coins
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

		sendTargets = append(sendTargets, sdk.Coin{
			Denom:  tokenDenom,
			Amount: tokenAmountOut,
		})
	}

	// process token transfer
	err = p.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		sender,
		sendTargets,
	)
	if err != nil {
		return err
	}

	// process lpToken transfer
	poolShare := lpService{
		denom:      lpToken.Denom,
		bankKeeper: p.bankKeeper,
	}
	if err := poolShare.pullPoolShare(ctx, sender, poolAmountIn); err != nil {
		return err
	}
	if err := poolShare.burnPoolShare(ctx, poolAmountIn); err != nil {
		return err
	}

	// save changes
	lpToken.TotalSupply = lpToken.TotalSupply.Sub(poolAmountIn)
	pool.Token = lpToken
	p.store.StorePool(ctx, pool)
	return nil
}
