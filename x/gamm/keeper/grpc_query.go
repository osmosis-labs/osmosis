package keeper

import (
	"context"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/cosmos/cosmos-sdk/store/prefix"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/c-osmosis/osmosis/x/gamm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	sdkIntMaxValue = sdk.NewInt(0)
)

func init() {
	maxInt := big.NewInt(2)
	maxInt = maxInt.Exp(maxInt, big.NewInt(255), nil)
	_sdkIntMaxValue, ok := sdk.NewIntFromString(maxInt.Sub(maxInt, big.NewInt(1)).String())
	if !ok {
		panic("Failed to calculate the max value of sdk.Int")
	}
	sdkIntMaxValue = _sdkIntMaxValue
}

var _ types.QueryServer = Keeper{}

func (k Keeper) Pool(
	ctx context.Context,
	req *types.QueryPoolRequest,
) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := k.GetPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	switch pool := pool.(type) {
	case *types.PoolAccount:
		return &types.QueryPoolResponse{Pool: *pool}, nil
	default:
		return nil, status.Error(codes.Internal, "invalid type of pool account")
	}
}

func (k Keeper) Pools(
	ctx context.Context,
	req *types.QueryPoolsRequest,
) (*types.QueryPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
	poolStore := prefix.NewStore(store, types.PaginationPoolNumbers)

	pools := []types.PoolAccount{}
	pageRes, err := query.Paginate(poolStore, req.Pagination, func(_, value []byte) error {
		poolId := sdk.BigEndianToUint64(value)
		poolI, err := k.GetPool(sdkCtx, poolId)

		if err != nil {
			return err
		}

		pool, ok := poolI.(*types.PoolAccount)
		if !ok {
			return fmt.Errorf("pool (%d) is not basic pool account", poolId)
		}

		pools = append(pools, *pool)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPoolsResponse{
		Pools:      pools,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) PoolParams(ctx context.Context, req *types.QueryPoolParamsRequest) (*types.QueryPoolParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := k.GetPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryPoolParamsResponse{
		Params: pool.GetPoolParams(),
	}, nil
}

func (k Keeper) TotalShare(ctx context.Context, req *types.QueryTotalShareRequest) (*types.QueryTotalShareResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := k.GetPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryTotalShareResponse{
		TotalShare: pool.GetTotalShare(),
	}, nil
}

func (k Keeper) Records(ctx context.Context, req *types.QueryRecordsRequest) (*types.QueryRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := k.GetPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryRecordsResponse{
		Records: pool.GetAllRecords(),
	}, nil
}

func (k Keeper) SpotPrice(ctx context.Context, req *types.QuerySpotPriceRequest) (*types.QuerySpotPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.TokenInDenom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	if req.TokenOutDenom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	var sp sdk.Dec
	var err error
	if req.SansSwapFee {
		sp, err = k.CalculateSpotPriceSansSwapFee(sdkCtx, req.PoolId, req.TokenInDenom, req.TokenOutDenom)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		sp, err = k.CalculateSpotPrice(sdkCtx, req.PoolId, req.TokenInDenom, req.TokenOutDenom)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &types.QuerySpotPriceResponse{
		SpotPrice: sp.String(),
	}, nil
}

func (k Keeper) EstimateSwapExactAmountIn(ctx context.Context, req *types.QuerySwapExactAmountInRequest) (*types.QuerySwapExactAmountInResponse, error) {
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

	tokenOutAmount, err := k.MultihopSwapExactAmountIn(sdkCtx, sender, req.Routes, tokenIn, sdk.NewInt(1))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySwapExactAmountInResponse{
		TokenOutAmount: tokenOutAmount,
	}, nil
}

func (k Keeper) EstimateSwapExactAmountOut(ctx context.Context, req *types.QuerySwapExactAmountOutRequest) (*types.QuerySwapExactAmountOutResponse, error) {
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

	tokenInAmount, err := k.MultihopSwapExactAmountOut(sdkCtx, sender, req.Routes, sdkIntMaxValue, tokenOut)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySwapExactAmountOutResponse{
		TokenInAmount: tokenInAmount,
	}, nil
}
