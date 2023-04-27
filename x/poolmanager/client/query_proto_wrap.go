package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/osmosis-labs/osmosis/v15/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/client/queryproto"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// This file should evolve to being code gen'd, off of `proto/poolmanager/v1beta/query.yml`

type Querier struct {
	K poolmanager.Keeper
}

func NewQuerier(k poolmanager.Keeper) Querier {
	return Querier{k}
}

func (q Querier) Params(ctx sdk.Context,
	req queryproto.ParamsRequest,
) (*queryproto.ParamsResponse, error) {
	params := q.K.GetParams(ctx)
	return &queryproto.ParamsResponse{Params: params}, nil
}

// EstimateSwapExactAmountIn estimates input token amount for a swap.
func (q Querier) EstimateSwapExactAmountIn(ctx sdk.Context, req queryproto.EstimateSwapExactAmountInRequest) (*queryproto.EstimateSwapExactAmountInResponse, error) {
	if req.TokenIn == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	tokenIn, err := sdk.ParseCoinNormalized(req.TokenIn)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token: %s", err.Error())
	}

	tokenOutAmount, err := q.K.MultihopEstimateOutGivenExactAmountIn(ctx, req.Routes, tokenIn)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.EstimateSwapExactAmountInResponse{
		TokenOutAmount: tokenOutAmount,
	}, nil
}

// EstimateSwapExactAmountOut estimates token output amount for a swap.
func (q Querier) EstimateSwapExactAmountOut(ctx sdk.Context, req queryproto.EstimateSwapExactAmountOutRequest) (*queryproto.EstimateSwapExactAmountOutResponse, error) {
	if req.TokenOut == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	if err := types.SwapAmountOutRoutes(req.Routes).Validate(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	tokenOut, err := sdk.ParseCoinNormalized(req.TokenOut)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token: %s", err.Error())
	}

	tokenInAmount, err := q.K.MultihopEstimateInGivenExactAmountOut(ctx, req.Routes, tokenOut)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.EstimateSwapExactAmountOutResponse{
		TokenInAmount: tokenInAmount,
	}, nil
}

func (q Querier) EstimateSinglePoolSwapExactAmountOut(ctx sdk.Context, req queryproto.EstimateSinglePoolSwapExactAmountOutRequest) (*queryproto.EstimateSwapExactAmountOutResponse, error) {
	routeReq := &queryproto.EstimateSwapExactAmountOutRequest{
		PoolId:   req.PoolId,
		TokenOut: req.TokenOut,
		Routes:   types.SwapAmountOutRoutes{{PoolId: req.PoolId, TokenInDenom: req.TokenInDenom}},
	}
	return q.EstimateSwapExactAmountOut(ctx, *routeReq)
}

func (q Querier) EstimateSinglePoolSwapExactAmountIn(ctx sdk.Context, req queryproto.EstimateSinglePoolSwapExactAmountInRequest) (*queryproto.EstimateSwapExactAmountInResponse, error) {
	routeReq := &queryproto.EstimateSwapExactAmountInRequest{
		PoolId:  req.PoolId,
		TokenIn: req.TokenIn,
		Routes:  types.SwapAmountInRoutes{{PoolId: req.PoolId, TokenOutDenom: req.TokenOutDenom}},
	}
	return q.EstimateSwapExactAmountIn(ctx, *routeReq)
}

// NumPools returns total number of pools.
func (q Querier) NumPools(ctx sdk.Context, _ queryproto.NumPoolsRequest) (*queryproto.NumPoolsResponse, error) {
	return &queryproto.NumPoolsResponse{
		NumPools: q.K.GetNextPoolId(ctx) - 1,
	}, nil
}

// Pool returns the pool specified by id.
func (q Querier) Pool(ctx sdk.Context, req queryproto.PoolRequest) (*queryproto.PoolResponse, error) {
	pool, err := q.K.GetPool(ctx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	any, err := codectypes.NewAnyWithValue(pool)
	if err != nil {
		return nil, err
	}

	return &queryproto.PoolResponse{
		Pool: any,
	}, nil
}

func (q Querier) AllPools(ctx sdk.Context, req queryproto.AllPoolsRequest) (*queryproto.AllPoolsResponse, error) {
	pools, err := q.K.AllPools(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var anyPools []*codectypes.Any
	for _, pool := range pools {
		any, err := codectypes.NewAnyWithValue(pool)
		if err != nil {
			return nil, err
		}
		anyPools = append(anyPools, any)
	}

	return &queryproto.AllPoolsResponse{
		Pools: anyPools,
	}, nil
}

// SpotPrice returns the spot price of the pool with the given quote and base asset denoms.
func (q Querier) SpotPrice(ctx sdk.Context, req queryproto.SpotPriceRequest) (*queryproto.SpotPriceResponse, error) {
	if req.BaseAssetDenom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid base asset denom")
	}

	if req.QuoteAssetDenom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid quote asset denom")
	}

	sp, err := q.K.RouteCalculateSpotPrice(ctx, req.PoolId, req.QuoteAssetDenom, req.BaseAssetDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.SpotPriceResponse{
		SpotPrice: sp.String(),
	}, err
}

// TotalPoolLiquidity returns the total liquidity of the pool.
func (q Querier) TotalPoolLiquidity(ctx sdk.Context, req queryproto.TotalPoolLiquidityRequest) (*queryproto.TotalPoolLiquidityResponse, error) {
	if req.PoolId == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Pool Id")
	}

	poolI, err := q.K.GetPool(ctx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	coins, err := q.K.GetTotalPoolLiquidity(ctx, poolI.GetId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.TotalPoolLiquidityResponse{
		Liquidity: coins,
	}, nil
}
