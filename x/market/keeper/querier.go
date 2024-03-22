package keeper

import (
	"context"

	"github.com/osmosis-labs/osmosis/v23/x/market/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over q
type querier struct {
	Keeper
}

// NewQuerier returns an implementation of the market QueryServer interface
// for the provided Keeper.
func NewQuerier(keeper Keeper) types.QueryServer {
	return &querier{Keeper: keeper}
}

var _ types.QueryServer = querier{}

// Params queries params of market module
func (q querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryParamsResponse{Params: q.GetParams(ctx)}, nil
}

// Swap queries for swap simulation
func (q querier) Swap(c context.Context, req *types.QuerySwapRequest) (*types.QuerySwapResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if err := sdk.ValidateDenom(req.AskDenom); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ask denom")
	}

	offerCoin, err := sdk.ParseCoinNormalized(req.OfferCoin)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	retCoin, err := q.simulateSwap(ctx, offerCoin, req.AskDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySwapResponse{ReturnCoin: retCoin}, nil
}

// TerraPoolDelta queries terra pool delta
func (q querier) TerraPoolDelta(c context.Context, _ *types.QueryTerraPoolDeltaRequest) (*types.QueryTerraPoolDeltaResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	terraPoolDelta := q.GetTerraPoolDelta(ctx)
	return &types.QueryTerraPoolDeltaResponse{TerraPoolDelta: terraPoolDelta}, nil
}
