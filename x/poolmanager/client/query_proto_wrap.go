package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/queryprotov2"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// This file should evolve to being code gen'd, off of `proto/poolmanager/v1beta/query.yml`

type Querier struct {
	K *poolmanager.Keeper
}

func NewQuerier(k *poolmanager.Keeper) Querier {
	return Querier{k}
}

// QuerierV2 defines a wrapper around the x/poolmanager keeper providing gRPC method
// handlers for v2 queries.
type QuerierV2 struct {
	K poolmanager.Keeper
}

func NewV2Querier(k poolmanager.Keeper) QuerierV2 {
	return QuerierV2{K: k}
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
		TokenOut: req.TokenOut,
		Routes:   types.SwapAmountOutRoutes{{TokenInDenom: req.TokenInDenom, PoolId: req.PoolId}},
	}
	return q.EstimateSwapExactAmountOut(ctx, *routeReq)
}

func (q Querier) EstimateSinglePoolSwapExactAmountIn(ctx sdk.Context, req queryproto.EstimateSinglePoolSwapExactAmountInRequest) (*queryproto.EstimateSwapExactAmountInResponse, error) {
	routeReq := &queryproto.EstimateSwapExactAmountInRequest{
		TokenIn: req.TokenIn,
		Routes:  types.SwapAmountInRoutes{{TokenOutDenom: req.TokenOutDenom, PoolId: req.PoolId}},
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

// ListPoolsByDenom returns a list of pools filtered by denom
func (q Querier) ListPoolsByDenom(ctx sdk.Context, req queryproto.ListPoolsByDenomRequest) (*queryproto.ListPoolsByDenomResponse, error) {
	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}
	pools, err := q.K.ListPoolsByDenom(ctx, req.Denom)
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

	return &queryproto.ListPoolsByDenomResponse{
		Pools: anyPools,
	}, nil
}

// SpotPrice returns the spot price of the pool with the given quote and base asset denoms. 18 decimals.
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
		// Note: truncation exists here to maintain backwards compatibility.
		// This query has historically had 18 decimals in response.
		SpotPrice: sp.Dec().String(),
	}, err
}

// SpotPriceV2 returns the spot price of the pool with the given quote and base asset denoms. 36 decimals.
func (q QuerierV2) SpotPriceV2(ctx sdk.Context, req queryprotov2.SpotPriceRequest) (*queryprotov2.SpotPriceResponse, error) {
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

	return &queryprotov2.SpotPriceResponse{
		// Note: that this is a BigDec yielding 36 decimals.
		SpotPrice: sp,
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

// TotalVolumeForPool returns the total volume of the pool.
func (q Querier) TotalVolumeForPool(ctx sdk.Context, req queryproto.TotalVolumeForPoolRequest) (*queryproto.TotalVolumeForPoolResponse, error) {
	totalVolume := q.K.GetTotalVolumeForPool(ctx, req.PoolId)

	return &queryproto.TotalVolumeForPoolResponse{
		Volume: totalVolume,
	}, nil
}

// TradingPairTakerFee returns the taker fee for the given trading pair
func (q Querier) TradingPairTakerFee(ctx sdk.Context, req queryproto.TradingPairTakerFeeRequest) (*queryproto.TradingPairTakerFeeResponse, error) {
	tradingPairTakerFee, err := q.K.GetTradingPairTakerFee(ctx, req.Denom_0, req.Denom_1)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.TradingPairTakerFeeResponse{
		TakerFee: tradingPairTakerFee,
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
		return nil, status.Error(codes.Internal, poolErr.Error())
	}

	spotPriceBigDec, err := swapModule.CalculateSpotPrice(ctx, req.PoolId, req.FromCoin.Denom, req.ToCoinDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert to normal Dec
	spotPrice := spotPriceBigDec.Dec()

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
				InputCoin:  sdk.NewCoin(req.FromCoin.Denom, osmomath.ZeroInt()),
				OutputCoin: sdk.NewCoin(req.ToCoinDenom, osmomath.ZeroInt()),
			}, nil
		}
	}

	// Process the estimates according to the pool type.
	switch poolI.GetType() {
	case types.Balancer:
		return q.K.EstimateTradeBasedOnPriceImpactBalancerPool(
			ctx, req, spotPrice, adjustedMaxPriceImpact, swapModule, poolI,
		)
	case types.Stableswap:
		return q.K.EstimateTradeBasedOnPriceImpactStableSwapPool(
			ctx, req, spotPrice, adjustedMaxPriceImpact, swapModule, poolI,
		)
	case types.Concentrated:
		return q.K.EstimateTradeBasedOnPriceImpactConcentratedLiquidity(
			ctx, req, spotPrice, adjustedMaxPriceImpact, swapModule, poolI,
		)
	default:
		return nil, status.Error(codes.Internal, "pool type not supported")
	}
}

func (q Querier) AllTakerFeeShareAgreements(ctx sdk.Context, req queryproto.AllTakerFeeShareAgreementsRequest) (*queryproto.AllTakerFeeShareAgreementsResponse, error) {
	takerFeeShareAgreements, err := q.K.GetAllTakerFeesShareAgreements(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &queryproto.AllTakerFeeShareAgreementsResponse{
		TakerFeeShareAgreements: takerFeeShareAgreements,
	}, nil
}

func (q Querier) TakerFeeShareAgreementFromDenom(ctx sdk.Context, req queryproto.TakerFeeShareAgreementFromDenomRequest) (*queryproto.TakerFeeShareAgreementFromDenomResponse, error) {
	takerFeeShareAgreement, found := q.K.GetTakerFeeShareAgreementFromDenomUNSAFE(req.Denom)
	if !found {
		return nil, status.Error(codes.NotFound, "taker fee share agreement not found")
	}
	return &queryproto.TakerFeeShareAgreementFromDenomResponse{
		TakerFeeShareAgreement: takerFeeShareAgreement,
	}, nil
}

func (q Querier) TakerFeeShareDenomsToAccruedValue(ctx sdk.Context, req queryproto.TakerFeeShareDenomsToAccruedValueRequest) (*queryproto.TakerFeeShareDenomsToAccruedValueResponse, error) {
	accruedValue, err := q.K.GetTakerFeeShareDenomsToAccruedValue(ctx, req.Denom, req.TakerFeeDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.TakerFeeShareDenomsToAccruedValueResponse{
		Amount: accruedValue,
	}, nil
}

func (q Querier) AllTakerFeeShareAccumulators(ctx sdk.Context, req queryproto.AllTakerFeeShareAccumulatorsRequest) (*queryproto.AllTakerFeeShareAccumulatorsResponse, error) {
	takerFeeSkimAccumulators, err := q.K.GetAllTakerFeeShareAccumulators(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.AllTakerFeeShareAccumulatorsResponse{
		TakerFeeSkimAccumulators: takerFeeSkimAccumulators,
	}, nil
}

func (q Querier) RegisteredAlloyedPoolFromDenom(ctx sdk.Context, req queryproto.RegisteredAlloyedPoolFromDenomRequest) (*queryproto.RegisteredAlloyedPoolFromDenomResponse, error) {
	contractState, found := q.K.GetRegisteredAlloyedPoolFromDenomUNSAFE(req.Denom)
	if !found {
		return nil, status.Error(codes.NotFound, "denom not found")
	}

	return &queryproto.RegisteredAlloyedPoolFromDenomResponse{
		ContractState: contractState,
	}, nil
}

func (q Querier) RegisteredAlloyedPoolFromPoolId(ctx sdk.Context, req queryproto.RegisteredAlloyedPoolFromPoolIdRequest) (*queryproto.RegisteredAlloyedPoolFromPoolIdResponse, error) {
	contractState, err := q.K.GetRegisteredAlloyedPoolFromPoolIdUNSAFE(ctx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "pool not found")
	}

	return &queryproto.RegisteredAlloyedPoolFromPoolIdResponse{
		ContractState: contractState,
	}, nil
}

func (q Querier) AllRegisteredAlloyedPools(ctx sdk.Context, req queryproto.AllRegisteredAlloyedPoolsRequest) (*queryproto.AllRegisteredAlloyedPoolsResponse, error) {
	contractStates, err := q.K.GetAllRegisteredAlloyedPools(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.AllRegisteredAlloyedPoolsResponse{
		ContractStates: contractStates,
	}, nil
}
