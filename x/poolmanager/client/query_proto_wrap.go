package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/osmosis-labs/osmosis/v17/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v17/x/poolmanager/client/queryproto"
	"github.com/osmosis-labs/osmosis/v17/x/poolmanager/types"
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

// EstimateSwapExactAmountInWithPrimitiveTypes runs same logic with EstimateSwapExactAmountIn
// but instead takes array of primitive types in the request to support query through grpc-gateway.
func (q Querier) EstimateSwapExactAmountInWithPrimitiveTypes(ctx sdk.Context, req queryproto.EstimateSwapExactAmountInWithPrimitiveTypesRequest) (*queryproto.EstimateSwapExactAmountInResponse, error) {
	if req.TokenIn == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	tokenIn, err := sdk.ParseCoinNormalized(req.TokenIn)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token: %s", err.Error())
	}

	var routes []types.SwapAmountInRoute

	for idx, poolId := range req.RoutesPoolId {
		var route types.SwapAmountInRoute
		route.PoolId = poolId
		route.TokenOutDenom = req.RoutesTokenOutDenom[idx]

		routes = append(routes, route)
	}

	tokenOutAmount, err := q.K.MultihopEstimateOutGivenExactAmountIn(ctx, routes, tokenIn)
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

// EstimateSwapExactAmountOut estimates token output amount for a swap.
func (q Querier) EstimateSwapExactAmountOutWithPrimitiveTypes(ctx sdk.Context, req queryproto.EstimateSwapExactAmountOutWithPrimitiveTypesRequest) (*queryproto.EstimateSwapExactAmountOutResponse, error) {
	if req.TokenOut == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	var routes []types.SwapAmountOutRoute

	for idx, poolId := range req.RoutesPoolId {
		var route types.SwapAmountOutRoute
		route.PoolId = poolId
		route.TokenInDenom = req.RoutesTokenInDenom[idx]
	}

	if err := types.SwapAmountOutRoutes(routes).Validate(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	tokenOut, err := sdk.ParseCoinNormalized(req.TokenOut)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token: %s", err.Error())
	}

	tokenInAmount, err := q.K.MultihopEstimateInGivenExactAmountOut(ctx, routes, tokenOut)
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
// based on external price and maximum price impact.
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

	// If ExternalPrice is specified we need to adjust the maxPriceImpact based on the deviation between spot and
	// external price.
	adjustedMaxPriceImpact := req.MaxPriceImpact
	if !req.ExternalPrice.IsZero() {
		priceDeviation := spotPrice.Sub(req.ExternalPrice).Quo(req.ExternalPrice)
		adjustedMaxPriceImpact = adjustedMaxPriceImpact.Sub(priceDeviation)

		// If the adjusted max price impact is negative or zero it means the difference between spot and external
		// already exceeds the max price impact.
		if adjustedMaxPriceImpact.IsZero() || adjustedMaxPriceImpact.IsNegative() {
			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  sdk.NewCoin(req.FromCoin.Denom, sdk.ZeroInt()),
				OutputCoin: sdk.NewCoin(req.ToCoinDenom, sdk.ZeroInt()),
			}, nil
		}
	}

	// Process the estimates according to the pool type.
	switch poolI.GetType() {
	case types.Balancer:
		return q.estimateTradeBasedOnPriceImpactBalancerPool(
			ctx, req, spotPrice, adjustedMaxPriceImpact, swapModule, poolI,
		)
	case types.Stableswap:
		return q.estimateTradeBasedOnPriceImpactStableSwapPool(
			ctx, req, spotPrice, adjustedMaxPriceImpact, swapModule, poolI,
		)
	case types.Concentrated:
		return q.estimateTradeBasedOnPriceImpactConcentratedLiquidity(
			ctx, req, spotPrice, adjustedMaxPriceImpact, swapModule, poolI,
		)
	default:
		return nil, status.Error(codes.Internal, "pool not supported")
	}
}

// estimateTradeBasedOnPriceImpactBalancerPool estimates a trade based on price impact for a balancer pool type.
// For a balancer pool if an amount entered is greater than the total pool liquidity the trade estimated would be
// the full liquidity of the other token. If the amount is small it would return a close 1:1 trade of the
// smallest units.
func (q Querier) estimateTradeBasedOnPriceImpactBalancerPool(
	ctx sdk.Context,
	req queryproto.EstimateTradeBasedOnPriceImpactRequest,
	spotPrice, adjustedMaxPriceImpact sdk.Dec,
	swapModule types.PoolModuleI,
	poolI types.PoolI,
) (*queryproto.EstimateTradeBasedOnPriceImpactResponse, error) {

	// There isn't a case where the tokenOut could be zero or an error is received but those possibilities are handled
	// anyway.
	tokenOut, err := swapModule.CalcOutAmtGivenIn(ctx, poolI, req.FromCoin, req.ToCoinDenom, sdk.ZeroDec())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if tokenOut.IsZero() {
		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  sdk.NewCoin(req.FromCoin.Denom, sdk.ZeroInt()),
			OutputCoin: sdk.NewCoin(req.ToCoinDenom, sdk.ZeroInt()),
		}, nil
	}

	// Validate if the trade as is respects the price impact, if it does re-estimate it with a swap fee and return
	// the result.
	currTradePrice := sdk.NewDec(req.FromCoin.Amount.Int64()).QuoInt(tokenOut.Amount)
	priceDeviation := currTradePrice.Sub(spotPrice).Quo(spotPrice).Abs()

	if priceDeviation.LTE(adjustedMaxPriceImpact) {
		tokenOut, err = swapModule.CalcOutAmtGivenIn(
			ctx, poolI, req.FromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx),
		)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  req.FromCoin,
			OutputCoin: tokenOut,
		}, nil
	}

	// Define low and high amount to search between. Start from 1 and req.FromCoin.Amount as initial range.
	lowAmount := sdk.OneInt()
	highAmount := req.FromCoin.Amount
	currFromCoin := req.FromCoin

	for lowAmount.LTE(highAmount) {
		// Calculate currFromCoin as the new middle amount to try trade.
		midAmount := lowAmount.Add(highAmount).Quo(sdk.NewInt(2))
		currFromCoin = sdk.NewCoin(req.FromCoin.Denom, midAmount)

		// There isn't a case where the tokenOut could be zero or an error is received but those possibilities are
		// handled anyway.
		tokenOut, err := swapModule.CalcOutAmtGivenIn(
			ctx, poolI, currFromCoin, req.ToCoinDenom, sdk.ZeroDec(),
		)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		if tokenOut.IsZero() {
			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  sdk.NewCoin(req.FromCoin.Denom, sdk.ZeroInt()),
				OutputCoin: sdk.NewCoin(req.ToCoinDenom, sdk.ZeroInt()),
			}, nil
		}

		currTradePrice := sdk.NewDec(currFromCoin.Amount.Int64()).QuoInt(tokenOut.Amount)
		priceDeviation := currTradePrice.Sub(spotPrice).Quo(spotPrice).Abs()

		if priceDeviation.LTE(adjustedMaxPriceImpact) {
			lowAmount = midAmount.Add(sdk.OneInt())
		} else {
			highAmount = midAmount.Sub(sdk.OneInt())
		}
	}

	// highAmount is 0 it means the loop has iterated to the end without finding a viable trade that respects
	// the price impact.
	if highAmount.IsZero() {
		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  sdk.NewCoin(req.FromCoin.Denom, sdk.ZeroInt()),
			OutputCoin: sdk.NewCoin(req.ToCoinDenom, sdk.ZeroInt()),
		}, nil
	}

	tokenOut, err = swapModule.CalcOutAmtGivenIn(
		ctx, poolI, currFromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
		InputCoin:  currFromCoin,
		OutputCoin: tokenOut,
	}, nil
}

// estimateTradeBasedOnPriceImpactStableSwapPool estimates a trade based on price impact for a stableswap pool type.
// For a stableswap pool if an amount entered is greater than the total pool liquidity the trade estimated would
// `panic`. If the amount is small it would return an error, in the case of a `panic` we should ignore it
// and keep attempting lower input amounts while if it's a normal error we should return an empty trade.
func (q Querier) estimateTradeBasedOnPriceImpactStableSwapPool(
	ctx sdk.Context,
	req queryproto.EstimateTradeBasedOnPriceImpactRequest,
	spotPrice, adjustedMaxPriceImpact sdk.Dec,
	swapModule types.PoolModuleI,
	poolI types.PoolI,
) (*queryproto.EstimateTradeBasedOnPriceImpactResponse, error) {

	var tokenOut sdk.Coin
	var err error
	err = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		tokenOut, err = swapModule.CalcOutAmtGivenIn(ctx, poolI, req.FromCoin, req.ToCoinDenom, sdk.ZeroDec())
		return err
	})

	// Find out if the error is because the amount is too large or too little. The calculation should error
	// if the amount is too small, and it should panic if the amount is too large. If the amount is too large
	// we want to continue to iterate to find attempt to find a smaller value.
	if err != nil && !strings.Contains(err.Error(), "panic") {
		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  sdk.NewCoin(req.FromCoin.Denom, sdk.ZeroInt()),
			OutputCoin: sdk.NewCoin(req.ToCoinDenom, sdk.ZeroInt()),
		}, nil
	} else if err == nil {
		// Validate if the trade as is respects the price impact, if it does re-estimate it with a swap fee and return
		// the result.
		currTradePrice := sdk.NewDec(req.FromCoin.Amount.Int64()).QuoInt(tokenOut.Amount)
		priceDeviation := currTradePrice.Sub(spotPrice).Quo(spotPrice).Abs()

		if priceDeviation.LTE(adjustedMaxPriceImpact) {
			tokenOut, err = swapModule.CalcOutAmtGivenIn(
				ctx, poolI, req.FromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx),
			)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}

			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  req.FromCoin,
				OutputCoin: tokenOut,
			}, nil
		}
	}

	// Define low and high amount to search between. Start from 1 and req.FromCoin.Amount as initial range.
	lowAmount := sdk.OneInt()
	highAmount := req.FromCoin.Amount
	currFromCoin := req.FromCoin

	for lowAmount.LTE(highAmount) {
		// Calculate currFromCoin as the new middle amount to try trade.
		midAmount := lowAmount.Add(highAmount).Quo(sdk.NewInt(2))
		currFromCoin = sdk.NewCoin(req.FromCoin.Denom, midAmount)

		err = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			tokenOut, err = swapModule.CalcOutAmtGivenIn(ctx, poolI, currFromCoin, req.ToCoinDenom, sdk.ZeroDec())
			return err
		})

		// If it returns an error without a panic it means the input has become too small and we should return.
		if err != nil && !strings.Contains(err.Error(), "panic") {
			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  sdk.NewCoin(req.FromCoin.Denom, sdk.ZeroInt()),
				OutputCoin: sdk.NewCoin(req.ToCoinDenom, sdk.ZeroInt()),
			}, nil
		} else if err != nil {
			// If there is an error that does contain a panic it means the amount is still too large,
			// and we should continue halving.
			highAmount = midAmount.Sub(sdk.OneInt())
		} else {
			currTradePrice := sdk.NewDec(currFromCoin.Amount.Int64()).QuoInt(tokenOut.Amount)
			priceDeviation := currTradePrice.Sub(spotPrice).Quo(spotPrice).Abs()

			if priceDeviation.LTE(adjustedMaxPriceImpact) {
				lowAmount = midAmount.Add(sdk.OneInt())
			} else {
				highAmount = midAmount.Sub(sdk.OneInt())
			}
		}
	}

	// highAmount is 0 it means the loop has iterated to the end without finding a viable trade that respects
	// the price impact.
	if highAmount.IsZero() {
		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  sdk.NewCoin(req.FromCoin.Denom, sdk.ZeroInt()),
			OutputCoin: sdk.NewCoin(req.ToCoinDenom, sdk.ZeroInt()),
		}, nil
	}

	tokenOut, err = swapModule.CalcOutAmtGivenIn(
		ctx, poolI, currFromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
		InputCoin:  currFromCoin,
		OutputCoin: tokenOut,
	}, nil
}

// estimateTradeBasedOnPriceImpactConcentratedLiquidity estimates a trade based on price impact for a concentrated
// liquidity pool type. For a concentrated liquidity pool if an amount entered is greater than the total pool liquidity
// the trade estimated would error. If the amount is small it would return tokenOut to be 0 in which case we should
// return an empty trade. If the estimate returns an error we should ignore it and continue attempting to estimate
// by halving the input.
func (q Querier) estimateTradeBasedOnPriceImpactConcentratedLiquidity(
	ctx sdk.Context,
	req queryproto.EstimateTradeBasedOnPriceImpactRequest,
	spotPrice, adjustedMaxPriceImpact sdk.Dec,
	swapModule types.PoolModuleI,
	poolI types.PoolI,
) (*queryproto.EstimateTradeBasedOnPriceImpactResponse, error) {

	tokenOut, err := swapModule.CalcOutAmtGivenIn(ctx, poolI, req.FromCoin, req.ToCoinDenom, sdk.ZeroDec())
	// If there was no error we attempt to validate if the output is below the adjustedMaxPriceImpact.
	if err == nil {

		// If the tokenOut was returned to be zero it means the amount being traded is too small. We ignore the
		// error output here as it could mean that the input is too large.
		if tokenOut.IsZero() {
			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  sdk.NewCoin(req.FromCoin.Denom, sdk.ZeroInt()),
				OutputCoin: sdk.NewCoin(req.ToCoinDenom, sdk.ZeroInt()),
			}, nil
		}

		currTradePrice := sdk.NewDec(req.FromCoin.Amount.Int64()).QuoInt(tokenOut.Amount)
		priceDeviation := currTradePrice.Sub(spotPrice).Quo(spotPrice).Abs()

		if priceDeviation.LTE(adjustedMaxPriceImpact) {
			tokenOut, err = swapModule.CalcOutAmtGivenIn(
				ctx, poolI, req.FromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx),
			)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}

			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  req.FromCoin,
				OutputCoin: tokenOut,
			}, nil
		}
	}

	// Define low and high amount to search between. Start from 1 and req.FromCoin.Amount as initial range.
	lowAmount := sdk.OneInt()
	highAmount := req.FromCoin.Amount
	currFromCoin := req.FromCoin

	for lowAmount.LTE(highAmount) {
		// Calculate currFromCoin as the new middle amount to try trade.
		midAmount := lowAmount.Add(highAmount).Quo(sdk.NewInt(2))
		currFromCoin = sdk.NewCoin(req.FromCoin.Denom, midAmount)

		tokenOut, err := swapModule.CalcOutAmtGivenIn(ctx, poolI, currFromCoin, req.ToCoinDenom, sdk.ZeroDec())
		if err == nil {
			// If the tokenOut was returned to be zero it means the amount being traded is too small. We ignore the
			// error output here as it could mean that the input is too large.
			if tokenOut.IsZero() {
				return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
					InputCoin:  sdk.NewCoin(req.FromCoin.Denom, sdk.ZeroInt()),
					OutputCoin: sdk.NewCoin(req.ToCoinDenom, sdk.ZeroInt()),
				}, nil
			}

			currTradePrice := sdk.NewDec(currFromCoin.Amount.Int64()).QuoInt(tokenOut.Amount)
			priceDeviation := currTradePrice.Sub(spotPrice).Quo(spotPrice).Abs()

			if priceDeviation.LTE(adjustedMaxPriceImpact) {
				lowAmount = midAmount.Add(sdk.OneInt())
			} else {
				highAmount = midAmount.Sub(sdk.OneInt())
			}
		} else {
			highAmount = midAmount.Sub(sdk.OneInt())
		}
	}

	// highAmount is 0 it means the loop has iterated to the end without finding a viable trade that respects
	// the price impact.
	if highAmount.IsZero() {
		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  sdk.NewCoin(req.FromCoin.Denom, sdk.ZeroInt()),
			OutputCoin: sdk.NewCoin(req.ToCoinDenom, sdk.ZeroInt()),
		}, nil
	}

	tokenOut, err = swapModule.CalcOutAmtGivenIn(
		ctx, poolI, currFromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
		InputCoin:  currFromCoin,
		OutputCoin: tokenOut,
	}, nil
}
