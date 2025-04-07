package keeper

import (
	"context"
	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/market/types"
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

// ExchangeRequirements returns the exchange requirements for the market module.
func (q querier) ExchangeRequirements(c context.Context, _ *types.QueryExchangeRequirementsRequest) (*types.QueryExchangeRequirementsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	resp := &types.QueryExchangeRequirementsResponse{}

	resp.ExchangeRequirements = q.getExchangeRates(ctx)
	total := osmomath.ZeroDec()
	for _, req := range resp.ExchangeRequirements {
		total = total.Add(req.BaseCurrency.Amount.ToLegacyDec().Mul(req.ExchangeRate))
	}
	resp.Total = sdk.NewCoin(appparams.BaseCoinUnit, total.TruncateInt())
	return resp, nil
}
