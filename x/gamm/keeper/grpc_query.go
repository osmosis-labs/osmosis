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

	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/v2types"
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

// QuerierV2 defines a wrapper around the x/gamm keeper providing gRPC method
// handlers for v2 queries.
type QuerierV2 struct {
	Keeper
}

func NewV2Querier(k Keeper) QuerierV2 {
	return QuerierV2{Keeper: k}
}

// Pool checks if a pool exists and their respective poolWeights.
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

// NumPools returns total number of pools.
func (q Querier) NumPools(ctx context.Context, _ *types.QueryNumPoolsRequest) (*types.QueryNumPoolsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryNumPoolsResponse{
		NumPools: q.Keeper.GetNextPoolId(sdkCtx) - 1,
	}, nil
}

func (q Querier) PoolType(ctx context.Context, req *types.QueryPoolTypeRequest) (*types.QueryPoolTypeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	poolType, err := q.Keeper.GetPoolType(sdkCtx, req.PoolId)

	return &types.QueryPoolTypeResponse{
		PoolType: poolType,
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
	pool, err := q.Keeper.getPoolForSwap(sdkCtx, req.PoolId)
	if err != nil {
		return nil, err
	}

	numShares, newLiquidity, err := pool.CalcJoinPoolShares(sdkCtx, req.TokensIn, pool.GetSwapFee(sdkCtx))
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
	res, err := q.Pools(ctx, &types.QueryPoolsRequest{
		Pagination: &query.PageRequest{},
	})
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err != nil {
		return nil, err
	}

	pools := res.Pools

	var response = []*codectypes.Any{}

	// set filters
	min_liquidity := req.MinLiquidity
	pool_type := req.PoolType
	checks_needed := 0
	// increase amount of needed checks for each filter by 1
	if min_liquidity != nil {
		checks_needed++
	}

	if pool_type != "" {
		checks_needed++
	}

	for _, p := range pools {
		var checks = 0
		var pool types.PoolI

		err := q.cdc.UnpackAny(p, &pool)
		if err != nil {
			return nil, sdkerrors.ErrUnpackAny
		}
		poolId := pool.GetId()

		// if liquidity specified in request
		if min_liquidity != nil {
			poolLiquidity := pool.GetTotalPoolLiquidity(sdkCtx)
			amount_of_denoms := 0
			check_amount := false
			check_denoms := false

			if poolLiquidity.IsAllGTE(min_liquidity) {
				check_amount = true
			}

			for _, req_coin := range min_liquidity {
				for _, coin := range poolLiquidity {
					if req_coin.Denom == coin.Denom {
						amount_of_denoms++
					}
				}
			}

			if amount_of_denoms == len(min_liquidity) {
				check_denoms = true
			}

			if check_amount && check_denoms {
				checks++
			}
		}

		// if pool type specified in request
		if pool_type != "" {
			poolType, err := q.GetPoolType(sdkCtx, poolId)
			if err != nil {
				return nil, types.ErrPoolNotFound
			}

			if poolType == pool_type {
				checks++
			} else {
				continue
			}
		}

		if checks == checks_needed {
			response = append(response, p)
		}
	}

	return &types.QueryPoolsWithFilterResponse{
		Pools: response,
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
		return nil, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "share ratio is zero or negative")
	}

	exitCoins, err := pool.CalcExitPoolCoinsFromShares(sdkCtx, req.ShareInAmount, exitFee)
	if err != nil {
		return nil, err
	}

	return &types.QueryCalcExitPoolCoinsFromSharesResponse{TokensOut: exitCoins}, nil
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
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnpackAny, errMsg)
	}
}

// TotalPoolLiquidity returns total liquidity in pool.
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

	sp, err := q.Keeper.CalculateSpotPrice(sdkCtx, req.PoolId, req.BaseAssetDenom, req.QuoteAssetDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySpotPriceResponse{
		SpotPrice: sp.String(),
	}, nil
}

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

	return &v2types.QuerySpotPriceResponse{
		SpotPrice: sp.String(),
	}, nil
}

// TotalLiquidity returns total liquidity across all pools.
func (q Querier) TotalLiquidity(ctx context.Context, _ *types.QueryTotalLiquidityRequest) (*types.QueryTotalLiquidityResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryTotalLiquidityResponse{
		Liquidity: q.Keeper.GetTotalLiquidity(sdkCtx),
	}, nil
}

// EstimateSwapExactAmountIn estimates input token amount for a swap.
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

// EstimateSwapExactAmountOut estimates token output amount for a swap.
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
