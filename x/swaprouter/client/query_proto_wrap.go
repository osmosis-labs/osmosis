package client

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v12/x/swaprouter"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/client/queryproto"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

// This file should evolve to being code gen'd, off of `proto/swaprouter/v1beta/query.yml`

type Querier struct {
	K swaprouter.Keeper
}

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

func (q Querier) Params(ctx sdk.Context,
	req queryproto.ParamsRequest,
) (*queryproto.ParamsResponse, error) {
	params := q.K.GetParams(ctx)
	return &queryproto.ParamsResponse{Params: params}, nil
}

// EstimateSwapExactAmountIn estimates input token amount for a swap.
func (q Querier) EstimateSwapExactAmountIn(ctx sdk.Context, req queryproto.EstimateSwapExactAmountInRequest) (*queryproto.EstimateSwapExactAmountInResponse, error) {
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

	tokenOutAmount, err := q.K.RouteExactAmountIn(ctx, sender, req.Routes, tokenIn, sdk.NewInt(1))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.EstimateSwapExactAmountInResponse{
		TokenOutAmount: tokenOutAmount,
	}, nil
}

// EstimateSwapExactAmountOut estimates token output amount for a swap.
func (q Querier) EstimateSwapExactAmountOut(ctx sdk.Context, req queryproto.EstimateSwapExactAmountOutRequest) (*queryproto.EstimateSwapExactAmountOutResponse, error) {
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

	tokenInAmount, err := q.K.RouteExactAmountOut(ctx, sender, req.Routes, sdkIntMaxValue, tokenOut)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.EstimateSwapExactAmountOutResponse{
		TokenInAmount: tokenInAmount,
	}, nil
}

// NumPools returns total number of pools.
func (q Querier) NumPools(ctx sdk.Context, _ queryproto.NumPoolsRequest) (*queryproto.NumPoolsResponse, error) {
	return &queryproto.NumPoolsResponse{
		NumPools: q.K.GetNextPoolId(ctx) - 1,
	}, nil
}
