package keeper

import (
	"context"
	"fmt"
	"math/big"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/types"
)

var sdkIntMaxValue = sdk.NewInt(0)

func init() {
	maxInt := big.NewInt(2)
	maxInt = maxInt.Exp(maxInt, big.NewInt(256), nil)

	_sdkIntMaxValue, ok := sdk.NewIntFromString(maxInt.Sub(maxInt, big.NewInt(1)).String())
	if !ok {
		panic("Failed to calculate the max value of sdk.Int")
	}

	sdkIntMaxValue = _sdkIntMaxValue
}

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/gamm keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

func (q Querier) Pool(
	ctx context.Context,
	req *types.QueryPoolRequest,
) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := q.Keeper.GetPoolAndPoke(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	any, err := codectypes.NewAnyWithValue(pool)
	if err != nil {
		return nil, err
	}

	return &types.QueryPoolResponse{Pool: any}, nil
}

func (q Querier) Pools(
	ctx context.Context,
	req *types.QueryPoolsRequest,
) (*types.QueryPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(q.Keeper.storeKey)
	poolStore := prefix.NewStore(store, types.KeyPrefixPools)

	var anys []*codectypes.Any
	pageRes, err := query.Paginate(poolStore, req.Pagination, func(_, value []byte) error {
		poolI, err := q.Keeper.UnmarshalPool(value)
		if err != nil {
			return err
		}

		// Use GetPoolAndPoke function because it runs PokeWeights
		poolI, err = q.Keeper.GetPoolAndPoke(sdkCtx, poolI.GetId())
		if err != nil {
			return err
		}

		// TODO: pools query should not be balancer specific
		pool, ok := poolI.(*balancer.Pool)
		if !ok {
			return fmt.Errorf("pool (%d) is not basic pool", pool.GetId())
		}

		any, err := codectypes.NewAnyWithValue(pool)
		if err != nil {
			return err
		}

		anys = append(anys, any)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPoolsResponse{
		Pools:      anys,
		Pagination: pageRes,
	}, nil
}

func (q Querier) NumPools(ctx context.Context, _ *types.QueryNumPoolsRequest) (*types.QueryNumPoolsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryNumPoolsResponse{
<<<<<<< HEAD
		NumPools: q.Keeper.GetNextPoolNumberAndIncrement(sdkCtx) - 1,
=======
		NumPools: q.Keeper.GetNextPoolId(sdkCtx) - 1,
>>>>>>> 8cac0906 (gamm keeper delete obsolete files (#2160))
	}, nil
}

func (q Querier) PoolParams(ctx context.Context, req *types.QueryPoolParamsRequest) (*types.QueryPoolParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := q.Keeper.GetPoolAndPoke(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	switch pool := pool.(type) {
	case *balancer.Pool:
		any, err := codectypes.NewAnyWithValue(&pool.PoolParams)
		if err != nil {
			return nil, err
		}

		return &types.QueryPoolParamsResponse{
			Params: any,
		}, nil

	default:
		errMsg := fmt.Sprintf("unrecognized %s pool type: %T", types.ModuleName, pool)
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnpackAny, errMsg)
	}
}

func (q Querier) TotalPoolLiquidity(ctx context.Context, req *types.QueryTotalPoolLiquidityRequest) (*types.QueryTotalPoolLiquidityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := q.Keeper.GetPoolAndPoke(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTotalPoolLiquidityResponse{
		Liquidity: pool.GetTotalPoolLiquidity(sdkCtx),
	}, nil
}

func (q Querier) TotalShares(ctx context.Context, req *types.QueryTotalSharesRequest) (*types.QueryTotalSharesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := q.Keeper.GetPoolAndPoke(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTotalSharesResponse{
		TotalShares: sdk.NewCoin(
			types.GetPoolShareDenom(req.PoolId),
			pool.GetTotalShares()),
	}, nil
}

func (q Querier) SpotPrice(ctx context.Context, req *types.QuerySpotPriceRequest) (*types.QuerySpotPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.BaseAssetDenom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid base asset denom")
	}

	if req.QuoteAssetDenom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid quote asset denom")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := q.Keeper.GetPoolAndPoke(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get pool by ID: %s", err)
	}

	sp, err := pool.SpotPrice(sdkCtx, req.BaseAssetDenom, req.QuoteAssetDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySpotPriceResponse{
		SpotPrice: sp.String(),
	}, nil
}

func (q Querier) TotalLiquidity(ctx context.Context, _ *types.QueryTotalLiquidityRequest) (*types.QueryTotalLiquidityResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryTotalLiquidityResponse{
		Liquidity: q.Keeper.GetTotalLiquidity(sdkCtx),
	}, nil
}

func (q Querier) EstimateSwapExactAmountIn(ctx context.Context, req *types.QuerySwapExactAmountInRequest) (*types.QuerySwapExactAmountInResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Sender == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	if req.TokenIn == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	if err := types.SwapAmountInRoutes(req.Routes).Validate(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	tokenIn, err := sdk.ParseCoinNormalized(req.TokenIn)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token: %s", err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	tokenOutAmount, err := q.Keeper.MultihopSwapExactAmountIn(sdkCtx, sender, req.Routes, tokenIn, sdk.NewInt(1))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySwapExactAmountInResponse{
		TokenOutAmount: tokenOutAmount,
	}, nil
}

func (q Querier) EstimateSwapExactAmountOut(ctx context.Context, req *types.QuerySwapExactAmountOutRequest) (*types.QuerySwapExactAmountOutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Sender == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	if req.TokenOut == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	if err := types.SwapAmountOutRoutes(req.Routes).Validate(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	tokenOut, err := sdk.ParseCoinNormalized(req.TokenOut)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token: %s", err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	tokenInAmount, err := q.Keeper.MultihopSwapExactAmountOut(sdkCtx, sender, req.Routes, sdkIntMaxValue, tokenOut)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySwapExactAmountOutResponse{
		TokenInAmount: tokenInAmount,
	}, nil
}
