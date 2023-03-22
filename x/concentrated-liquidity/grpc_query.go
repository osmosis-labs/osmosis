package concentrated_liquidity

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	clquery "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/query"
)

const (
	liquidityDepthRangeQueryLimit = 10000
)

var (
	liquidityDepthRangeQueryLimitInt = sdk.NewInt(liquidityDepthRangeQueryLimit)
)

var _ clquery.QueryServer = Querier{}

// Querier defines a wrapper around the x/concentrated-liquidity keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// UserPositions returns positions of a specified address
func (q Querier) UserPositions(ctx context.Context, req *clquery.QueryUserPositionsRequest) (*clquery.QueryUserPositionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	userPositions, err := q.Keeper.GetUserPositions(sdkCtx, sdkAddr, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	positions := make([]model.PositionWithUnderlyingAssetBreakdown, 0, len(userPositions))

	for _, position := range userPositions {
		// get the pool from the position
		pool, err := q.Keeper.getPoolById(sdkCtx, position.PoolId)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		asset0, asset1, err := CalculateUnderlyingAssetsFromPosition(sdkCtx, position, pool)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Append the position and underlying assets to the positions slice
		positions = append(positions, model.PositionWithUnderlyingAssetBreakdown{
			Position: position,
			Asset0:   asset0,
			Asset1:   asset1,
		})
	}

	return &clquery.QueryUserPositionsResponse{
		Positions: positions,
	}, nil
}

// Pools returns all concentrated pools in existence.
func (q Querier) Pools(
	ctx context.Context,
	req *clquery.QueryPoolsRequest,
) (*clquery.QueryPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(q.Keeper.storeKey)
	poolStore := prefix.NewStore(store, types.PoolPrefix)

	var anys []*codectypes.Any
	pageRes, err := query.Paginate(poolStore, req.Pagination, func(key, _ []byte) error {
		pool := model.Pool{}
		// Get the next pool from the poolStore and pass it to the pool variable
		_, err := osmoutils.Get(poolStore, key, &pool)
		if err != nil {
			return err
		}

		// Retrieve the poolInterface from the respective pool
		poolI, err := q.Keeper.GetPool(sdkCtx, pool.GetId())
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

	return &clquery.QueryPoolsResponse{
		Pools:      anys,
		Pagination: pageRes,
	}, nil
}

// Params returns module params
func (q Querier) Params(goCtx context.Context, req *clquery.QueryParamsRequest) (*clquery.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &clquery.QueryParamsResponse{Params: q.Keeper.GetParams(ctx)}, nil
}

// LiquidityDepthsForRange returns liquidity depths for the given range.
func (q Querier) LiquidityDepthsForRange(goCtx context.Context, req *clquery.QueryLiquidityDepthsForRangeRequest) (*clquery.QueryLiquidityDepthsForRangeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.LowerTick.GT(req.UpperTick) {
		return nil, types.InvalidLowerUpperTickError{LowerTick: req.LowerTick.Int64(), UpperTick: req.UpperTick.Int64()}
	}

	requestedRange := req.UpperTick.Sub(req.LowerTick)
	// use constant pre-defined to limit range and check if reuested range does not exceed max range
	if requestedRange.GT(liquidityDepthRangeQueryLimitInt) {
		return nil, types.QueryRangeUnsupportedError{RequestedRange: requestedRange, MaxRange: liquidityDepthRangeQueryLimitInt}
	}

	liquidityDepths, err := q.Keeper.GetPerTickLiquidityDepthFromRange(ctx, req.PoolId, req.LowerTick.Int64(), req.UpperTick.Int64())
	if err != nil {
		return nil, err
	}

	return &clquery.QueryLiquidityDepthsForRangeResponse{
		LiquidityDepths: liquidityDepths,
	}, nil
}

// TickLiquidityInBatches returns array of liquidity depths in the given range of lower tick and upper tick.
// Note that the space between the ticks in the returned array would always be guaranteed spacing greater than given batch unit.
func (q Querier) TotalLiquidityForRange(goCtx context.Context, req *clquery.QueryTotalLiquidityForRangeRequest) (*clquery.QueryTotalLiquidityForRangeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	liquidity, err := q.Keeper.GetTickLiquidityForRange(
		ctx,
		req.PoolId,
	)
	if err != nil {
		return nil, err
	}

	return &clquery.QueryTotalLiquidityForRangeResponse{Liquidity: liquidity}, nil
}

func (q Querier) ClaimableFees(ctx context.Context, req *clquery.QueryClaimableFeesRequest) (*clquery.QueryClaimableFeesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkAddr, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	claimableFees, err := q.Keeper.queryClaimableFees(sdkCtx, req.PoolId, sdkAddr, req.LowerTick, req.UpperTick)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &clquery.QueryClaimableFeesResponse{
		ClaimableFees: claimableFees,
	}, nil
}
