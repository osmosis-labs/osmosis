package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/v2types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/gamm keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// QuerierV2 defines a wrapper around the x/gamm keeper providing gRPC method
// handlers for v2 queries.
type QuerierV2 struct {
	Keeper
}

func NewV2Querier(k Keeper) QuerierV2 {
	return QuerierV2{Keeper: k}
}

// Pool checks if a pool exists and their respective poolWeights.
// Deprecated: use x/poolmanager's Pool query.
// nolint: staticcheck
func (q Querier) Pool(
	ctx context.Context,
	req *types.QueryPoolRequest,
) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// GetPool gets pool from poolmanager that has the knowledge of all pool ids
	// within Osmosis.
	pool, err := q.Keeper.poolManager.GetPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	any, err := codectypes.NewAnyWithValue(pool)
	if err != nil {
		return nil, err
	}

	// Deprecated: use x/poolmanager's Pool query.
	// nolint: staticcheck
	return &types.QueryPoolResponse{Pool: any}, nil
}

// Pools checks existence of multiple pools and their poolWeights
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

		any, err := codectypes.NewAnyWithValue(poolI)
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

// This query has been deprecated and has been moved to poolmanager module.
// nolint: staticcheck
func (q Querier) NumPools(ctx context.Context, _ *types.QueryNumPoolsRequest) (*types.QueryNumPoolsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryNumPoolsResponse{
		NumPools: q.poolManager.GetNextPoolId(sdkCtx) - 1,
	}, nil
}

func (q Querier) PoolType(ctx context.Context, req *types.QueryPoolTypeRequest) (*types.QueryPoolTypeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	poolType, err := q.Keeper.GetPoolType(sdkCtx, req.PoolId)

	poolTypeStr, ok := poolmanagertypes.PoolType_name[int32(poolType)]
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid pool type: %d", poolType)
	}

	return &types.QueryPoolTypeResponse{
		PoolType: poolTypeStr,
	}, err
}

// CalcJoinPoolShares queries the amount of shares you get by providing specific amount of tokens
func (q Querier) CalcJoinPoolShares(ctx context.Context, req *types.QueryCalcJoinPoolSharesRequest) (*types.QueryCalcJoinPoolSharesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.TokensIn == nil {
		return nil, status.Error(codes.InvalidArgument, "no tokens in")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	pool, err := q.Keeper.GetCFMMPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, err
	}

	numShares, newLiquidity, err := pool.CalcJoinPoolShares(sdkCtx, req.TokensIn, pool.GetSpreadFactor(sdkCtx))
	if err != nil {
		return nil, err
	}

	return &types.QueryCalcJoinPoolSharesResponse{
		ShareOutAmount: numShares,
		TokensOut:      newLiquidity,
	}, nil
}

// PoolsWithFilter query allows to query pools with specific parameters
func (q Querier) PoolsWithFilter(ctx context.Context, req *types.QueryPoolsWithFilterRequest) (*types.QueryPoolsWithFilterResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(q.Keeper.storeKey)
	poolStore := prefix.NewStore(store, types.KeyPrefixPools)
	minLiquidity, err := sdk.ParseCoinsNormalized(req.MinLiquidity)
	if err != nil {
		return nil, err
	}

	response := []*codectypes.Any{}
	pageRes, err := query.FilteredPaginate(poolStore, req.Pagination, func(_, value []byte, accumulate bool) (bool, error) {
		pool, err := q.Keeper.UnmarshalPool(value)
		if err != nil {
			return false, err
		}

		poolId := pool.GetId()

		// if liquidity specified in request
		if len(minLiquidity) > 0 {
			poolLiquidity := pool.GetTotalPoolLiquidity(sdkCtx)

			if !poolLiquidity.IsAllGTE(minLiquidity) {
				return false, nil
			}
		}

		// if pool type specified in request
		if req.PoolType != "" {
			poolType, err := q.GetPoolType(sdkCtx, poolId)
			if err != nil {
				return false, types.ErrPoolNotFound
			}

			poolTypeStr, ok := poolmanagertypes.PoolType_name[int32(poolType)]
			if !ok {
				return false, fmt.Errorf("%d pool type not found", int32(poolType))
			}

			if poolTypeStr != req.PoolType {
				return false, nil
			}
		}

		any, err := codectypes.NewAnyWithValue(pool)
		if err != nil {
			return false, err
		}

		if accumulate {
			response = append(response, any)
		}

		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPoolsWithFilterResponse{
		Pools:      response,
		Pagination: pageRes,
	}, nil
}

// CalcExitPoolCoinsFromShares queries the amount of tokens you get by exiting a specific amount of shares
func (q Querier) CalcExitPoolCoinsFromShares(ctx context.Context, req *types.QueryCalcExitPoolCoinsFromSharesRequest) (*types.QueryCalcExitPoolCoinsFromSharesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	pool, err := q.Keeper.GetPoolAndPoke(sdkCtx, req.PoolId)
	if err != nil {
		return nil, types.ErrPoolNotFound
	}

	exitFee := pool.GetExitFee(sdkCtx)

	totalSharesAmount := pool.GetTotalShares()
	if req.ShareInAmount.GTE(totalSharesAmount) || req.ShareInAmount.LTE(sdk.ZeroInt()) {
		return nil, errorsmod.Wrapf(types.ErrInvalidMathApprox, "share ratio is zero or negative")
	}

	exitCoins, err := pool.CalcExitPoolCoinsFromShares(sdkCtx, req.ShareInAmount, exitFee)
	if err != nil {
		return nil, err
	}

	return &types.QueryCalcExitPoolCoinsFromSharesResponse{TokensOut: exitCoins}, nil
}

// CalcJoinPoolNoSwapShares returns the amount of shares you'd get if joined a pool without a swap and tokens which need to be provided
func (q Querier) CalcJoinPoolNoSwapShares(ctx context.Context, req *types.QueryCalcJoinPoolNoSwapSharesRequest) (*types.QueryCalcJoinPoolNoSwapSharesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	pool, err := q.GetPoolAndPoke(sdkCtx, req.PoolId)
	if err != nil {
		return nil, err
	}

	sharesOut, tokensJoined, err := pool.CalcJoinPoolNoSwapShares(sdkCtx, req.TokensIn, pool.GetSpreadFactor(sdkCtx))
	if err != nil {
		return nil, err
	}

	return &types.QueryCalcJoinPoolNoSwapSharesResponse{
		TokensOut: tokensJoined,
		SharesOut: sharesOut,
	}, nil
}

// PoolParams queries a specified pool for its params.
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
		return nil, errorsmod.Wrap(sdkerrors.ErrUnpackAny, errMsg)
	}
}

// TotalPoolLiquidity returns total liquidity in pool.
// Deprecated: please use the alternative in x/poolmanager
// nolint: staticcheck
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

// TotalShares returns total pool shares.
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

// SpotPrice returns target pool asset prices on base and quote assets.
// nolint: staticcheck
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

	// Note: the base and quote asset argument order is intentionally incorrect
	// due to a historic bug in the original implementation.
	sp, err := q.Keeper.CalculateSpotPrice(sdkCtx, req.PoolId, req.BaseAssetDenom, req.QuoteAssetDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySpotPriceResponse{
		SpotPrice: sp.String(),
	}, nil
}

// Deeprecated: use alternate in x/poolmanager
// nolint: staticcheck
func (q QuerierV2) SpotPrice(ctx context.Context, req *v2types.QuerySpotPriceRequest) (*v2types.QuerySpotPriceResponse, error) {
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

	sp, err := q.Keeper.CalculateSpotPrice(sdkCtx, req.PoolId, req.QuoteAssetDenom, req.BaseAssetDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Deeprecated: use alternate in x/poolmanager
	// nolint: staticcheck
	return &v2types.QuerySpotPriceResponse{
		SpotPrice: sp.String(),
	}, nil
}

// TotalLiquidity returns total liquidity across all gamm pools.
func (q Querier) TotalLiquidity(ctx context.Context, _ *types.QueryTotalLiquidityRequest) (*types.QueryTotalLiquidityResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	totalLiquidity, err := q.Keeper.GetTotalLiquidity(sdkCtx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTotalLiquidityResponse{
		Liquidity: totalLiquidity,
	}, nil
}

// EstimateSwapExactAmountIn estimates input token amount for a swap.
// This query is deprecated and has been moved to poolmanager module.
// nolint: staticcheck
func (q Querier) EstimateSwapExactAmountIn(ctx context.Context, req *types.QuerySwapExactAmountInRequest) (*types.QuerySwapExactAmountInResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.TokenIn == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	tokenIn, err := sdk.ParseCoinNormalized(req.TokenIn)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token: %s", err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	tokenOutAmount, err := q.Keeper.poolManager.MultihopEstimateOutGivenExactAmountIn(sdkCtx, req.Routes, tokenIn)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySwapExactAmountInResponse{
		TokenOutAmount: tokenOutAmount,
	}, nil
}

// EstimateSwapExactAmountOut estimates token output amount for a swap.
// This query is deprecated and has been moved to poolmanager module.
// nolint: staticcheck
func (q Querier) EstimateSwapExactAmountOut(ctx context.Context, req *types.QuerySwapExactAmountOutRequest) (*types.QuerySwapExactAmountOutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.TokenOut == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	tokenOut, err := sdk.ParseCoinNormalized(req.TokenOut)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token: %s", err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	tokenInAmount, err := q.Keeper.poolManager.MultihopEstimateInGivenExactAmountOut(sdkCtx, req.Routes, tokenOut)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySwapExactAmountOutResponse{
		TokenInAmount: tokenInAmount,
	}, nil
}

// ConcentratedPoolIdLinkFromCFMM queries the concentrated pool id linked to a cfmm pool id.
func (q Querier) ConcentratedPoolIdLinkFromCFMM(ctx context.Context, req *types.QueryConcentratedPoolIdLinkFromCFMMRequest) (*types.QueryConcentratedPoolIdLinkFromCFMMResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.CfmmPoolId == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid cfmm pool id")
	}
	poolIdEntering, err := q.Keeper.GetLinkedConcentratedPoolID(sdk.UnwrapSDKContext(ctx), req.CfmmPoolId)
	if err != nil {
		return nil, err
	}

	return &types.QueryConcentratedPoolIdLinkFromCFMMResponse{
		ConcentratedPoolId: poolIdEntering,
	}, nil
}

func (q Querier) CFMMConcentratedPoolLinks(ctx context.Context, req *types.QueryCFMMConcentratedPoolLinksRequest) (*types.QueryCFMMConcentratedPoolLinksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	poolLinks, err := q.Keeper.GetAllMigrationInfo(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		return nil, err
	}

	return &types.QueryCFMMConcentratedPoolLinksResponse{
		MigrationRecords: &poolLinks,
	}, nil
}
