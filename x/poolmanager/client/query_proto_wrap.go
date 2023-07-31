package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/osmosis-labs/osmosis/v16/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v16/x/poolmanager/client/queryproto"
	"github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
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
	pool = pool.AsSerializablePool()

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
		any, err := codectypes.NewAnyWithValue(pool.AsSerializablePool())
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

// TotalLiquidity returns the total liquidity across all pools.
func (q Querier) TotalLiquidity(ctx sdk.Context, req queryproto.TotalLiquidityRequest) (*queryproto.TotalLiquidityResponse, error) {
	totalLiquidity, err := q.K.TotalLiquidity(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &queryproto.TotalLiquidityResponse{
		Liquidity: totalLiquidity,
	}, nil
}

// EstimateTradeBasedOnPriceImpact returns the input and output amount of coins for a pool trade
// based on twap value and maximum price impact.
func (q Querier) EstimateTradeBasedOnPriceImpact(
	ctx sdk.Context,
	req queryproto.EstimateTradeBasedOnPriceImpactRequest,
) (*queryproto.EstimateTradeBasedOnPriceImpactResponse, error) {
	if req.PoolId == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Pool Id")
	}

	if req.FromCoin.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid from coin denom")
	}

	if req.ToCoinDenom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid to coin denom")
	}

	swapModule, err := q.K.GetPoolModule(ctx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	poolI, poolErr := swapModule.GetPool(ctx, req.PoolId)
	if poolErr != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	spotPrice, err := swapModule.CalculateSpotPrice(ctx, req.PoolId, req.FromCoin.Denom, req.ToCoinDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// If TwapPrice is specified we need to adjust the maxPriceImpact based on the deviation between spot and twap.
	adjustedMaxPriceImpact := req.MaxPriceImpact
	if !req.TwapPrice.IsZero() {
		priceDeviation := spotPrice.Sub(req.TwapPrice).Quo(req.TwapPrice)
		adjustedMaxPriceImpact = adjustedMaxPriceImpact.Sub(priceDeviation)

		// If price deviation is greater than the adjustedMaxPriceImpact it means spot is greater than twap
		// therefore return 0 input and 0 output.
		if priceDeviation.GT(adjustedMaxPriceImpact) {
			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  sdk.NewCoin(req.FromCoin.Denom, sdk.NewInt(0)),
				OutputCoin: sdk.NewCoin(req.ToCoinDenom, sdk.NewInt(0)),
			}, nil
		}
	}

	// First, try the full 'from coin' amount
	tokenOut, err := swapModule.CalcOutAmtGivenIn(ctx, poolI, req.FromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// If the calculated amount of tokenOut is 0 it means that input value of fromCoin was too low.
	if tokenOut.IsZero() {
		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  req.FromCoin,
			OutputCoin: tokenOut,
		}, nil
	}

	currTradePrice := sdk.NewDec(tokenOut.Amount.Int64()).QuoInt(req.FromCoin.Amount)
	priceDeviation := currTradePrice.Sub(spotPrice).Quo(spotPrice).Abs()

	if priceDeviation.LTE(adjustedMaxPriceImpact) {
		// If the full 'from coin' amount results in a price deviation less than or equal to the adjusted max price
		// impact, return it
		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  req.FromCoin,
			OutputCoin: tokenOut,
		}, nil
	}

	// Define low and high amount to search between. Start from 1 and req.FromCoin.Amount as initial range.
	lowAmount := sdk.NewInt(1)
	highAmount := req.FromCoin.Amount

	for lowAmount.LTE(highAmount) {
		// Calculate middle amount
		midAmount := lowAmount.Add(highAmount).Quo(sdk.NewInt(2))

		// Update currFromCoin
		currFromCoin := sdk.NewCoin(req.FromCoin.Denom, midAmount)

		tokenOut, err := swapModule.CalcOutAmtGivenIn(
			ctx, poolI, currFromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx))
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// If the calculated amount of tokenOut is 0 it means that input value of fromCoin was too low.
		if tokenOut.IsZero() {
			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  req.FromCoin,
				OutputCoin: tokenOut,
			}, nil
		}

		currTradePrice := sdk.NewDec(tokenOut.Amount.Int64()).QuoInt(currFromCoin.Amount)
		priceDeviation := currTradePrice.Sub(spotPrice).Quo(spotPrice).Abs()

		// Check priceDeviation against adjustedMaxPriceImpact
		if priceDeviation.LTE(adjustedMaxPriceImpact) {
			lowAmount = midAmount.Add(sdk.NewInt(1))
		} else {
			highAmount = midAmount.Sub(sdk.NewInt(1))
		}
	}

	// If lowAmount exceeds highAmount, then binary search ends and the last successful trade amount is `highAmount`
	finalTradeAmount := sdk.NewCoin(req.FromCoin.Denom, highAmount)

	// Calculate final tokenOut using highAmount
	tokenOut, err = swapModule.CalcOutAmtGivenIn(
		ctx, poolI, finalTradeAmount, req.ToCoinDenom, poolI.GetSpreadFactor(ctx),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Return the result
	return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
		InputCoin:  finalTradeAmount,
		OutputCoin: tokenOut,
	}, nil
}
