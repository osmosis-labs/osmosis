package pool

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type Service interface {
	CreatePool(sdk.Context, sdk.AccAddress, sdk.Dec, []types.TokenInfo) error
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

func (p poolService) CreatePool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	swapFee sdk.Dec,
	tokenInfo []types.TokenInfo,
) error {
	if len(tokenInfo) < 2 {
		return sdkerrors.Wrapf(
			types.ErrInvalidRequest,
			"token info length should be at least 2",
		)
	}

	records := make(map[string]types.Record, len(tokenInfo))
	for _, info := range tokenInfo {
		records[info.Denom] = types.Record{
			DenormalizedWeight: info.Ratio,
			Balance:            info.Amount,
		}
	}

	pool := types.Pool{
		Id:          p.store.GetNextPoolNumber(ctx),
		SwapFee:     swapFee,
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

	var sendTargets sdk.Coins
	for _, maxAmountIn := range maxAmountsIn {
		var (
			tokenDenom    = maxAmountIn.Denom
			record        = pool.Records[tokenDenom]
			tokenAmountIn = poolRatio.Mul(record.Balance.ToDec()).TruncateInt()
		)
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

	err = p.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		sender,
		types.ModuleName,
		sendTargets,
	)
	if err != nil {
		return err
	}

	poolShare := lpService{
		denom:      lpToken.Name,
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

	poolShare := lpService{
		denom:      lpToken.Name,
		bankKeeper: p.bankKeeper,
	}
	if err := poolShare.pullPoolShare(ctx, sender, poolAmountIn); err != nil {
		return err
	}
	if err := poolShare.burnPoolShare(ctx, poolAmountIn); err != nil {
		return err
	}

	var sendTargets sdk.Coins
	for _, minAmountOut := range minAmountsOut {
		var (
			tokenDenom     = minAmountOut.Denom
			record         = pool.Records[tokenDenom]
			tokenAmountOut = poolRatio.Mul(record.Balance.ToDec()).TruncateInt()
		)
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

	err = p.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		sender,
		sendTargets,
	)
	if err != nil {
		return err
	}

	// save changes
	lpToken.TotalSupply = lpToken.TotalSupply.Sub(poolAmountIn)
	pool.Token = lpToken
	p.store.StorePool(ctx, pool)
	return nil
}
