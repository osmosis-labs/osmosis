package pool

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type (
	Service interface {
		// Viewer
		GetPool(sdk.Context, uint64) (types.Pool, error)
		// TODO: handle the pagination. For now, just returns the all pools.
		GetPools(sdk.Context) ([]types.Pool, error)
		GetSwapFee(sdk.Context, uint64) (sdk.Dec, error)
		GetShareInfo(sdk.Context, uint64) (types.LP, error)
		GetTokenBalance(sdk.Context, uint64) (sdk.Coins, error)
		GetSpotPrice(sdk.Context, uint64, string, string) (sdk.Int, error)

		// Sender
		LiquidityPoolTransactor
		LiquiditySwapTransactor
	}
)

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

func (p poolService) GetPools(ctx sdk.Context) ([]types.Pool, error) {
	pools, err := p.store.FetchAllPools(ctx)
	if err != nil {
		return nil, err
	}
	return pools, nil
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
