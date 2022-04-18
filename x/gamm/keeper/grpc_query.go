package keeper

import (
	"context"
	"fmt"
	"math/big"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/cosmos/cosmos-sdk/store/prefix"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

var (
	sdkIntMaxValue = sdk.NewInt(0)
)

func init() {
	maxInt := big.NewInt(2)
	maxInt = maxInt.Exp(maxInt, big.NewInt(256), nil)
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
	any, err := codectypes.NewAnyWithValue(pool)
	if err != nil {
		return nil, err
	}
	return &types.QueryPoolResponse{Pool: any}, nil
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
	poolStore := prefix.NewStore(store, types.KeyPrefixPools)

	var anys []*codectypes.Any
	pageRes, err := query.Paginate(poolStore, req.Pagination, func(_, value []byte) error {
		poolI, err := k.UnmarshalPool(value)
		if err != nil {
			return err
		}

		// Use GetPool function because it runs PokeWeights
		poolI, err = k.GetPool(sdkCtx, poolI.GetId())
		if err != nil {
			return err
		}

		// TODO: pools query should not be balancer specific
		pool, ok := poolI.(*balancer.Pool)
		if !ok {
			return fmt.Errorf("pool (%d) is not basic pool", pool.GetId())
		}

		any, err := codectypes.NewAnyWithValue(pool)
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

func (k Keeper) NumPools(
	ctx context.Context,
	req *types.QueryNumPoolsRequest,
) (*types.QueryNumPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryNumPoolsResponse{
		NumPools: k.GetNextPoolNumberAndIncrement(sdkCtx) - 1,
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

func (k Keeper) TotalShares(ctx context.Context, req *types.QueryTotalSharesRequest) (*types.QueryTotalSharesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := k.GetPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryTotalSharesResponse{
		TotalShares: pool.GetTotalShares(),
	}, nil
}

func (k Keeper) PoolAssets(ctx context.Context, req *types.QueryPoolAssetsRequest) (*types.QueryPoolAssetsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := k.GetPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryPoolAssetsResponse{
		PoolAssets: pool.GetAllPoolAssets(),
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

	// Return the spot price anyway, even if the pool is inactive.

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	var sp sdk.Dec
	var err error
	if req.WithSwapFee {
		sp, err = k.CalculateSpotPriceWithSwapFee(sdkCtx, req.PoolId, req.TokenInDenom, req.TokenOutDenom)
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

func (k Keeper) TotalLiquidity(ctx context.Context, req *types.QueryTotalLiquidityRequest) (*types.QueryTotalLiquidityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryTotalLiquidityResponse{
		Liquidity: k.GetTotalLiquidity(sdkCtx),
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
