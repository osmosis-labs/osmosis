package client

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	clquery "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/client/queryproto"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
)

// Querier defines a wrapper around the x/concentrated-liquidity keeper providing gRPC method
// handlers.
type Querier struct {
	cl.Keeper
}

func NewQuerier(k cl.Keeper) Querier {
	return Querier{Keeper: k}
}

// UserPositions returns positions of a specified address
func (q Querier) UserPositions(ctx sdk.Context, req clquery.UserPositionsRequest) (*clquery.UserPositionsResponse, error) {
	sdkAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	userPositions, err := q.Keeper.GetUserPositions(ctx, sdkAddr, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	positions := make([]model.PositionWithUnderlyingAssetBreakdown, 0, len(userPositions))

	for _, position := range userPositions {
		// get the pool from the position
		pool, err := q.Keeper.GetPoolFromPoolIdAndConvertToConcentrated(ctx, position.PoolId)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		asset0, asset1, err := cl.CalculateUnderlyingAssetsFromPosition(ctx, position, pool)
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

	return &clquery.UserPositionsResponse{
		Positions: positions,
	}, nil
}

// PositionById returns a position with the specified id.
func (q Querier) PositionById(ctx sdk.Context, req clquery.PositionByIdRequest) (*clquery.PositionByIdResponse, error) {
	position, err := q.Keeper.GetPosition(ctx, req.PositionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	positionPool, err := q.Keeper.GetPoolFromPoolIdAndConvertToConcentrated(ctx, position.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	asset0, asset1, err := cl.CalculateUnderlyingAssetsFromPosition(ctx, position, positionPool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &clquery.PositionByIdResponse{
		Position: model.PositionWithUnderlyingAssetBreakdown{
			Position: position,
			Asset0:   asset0,
			Asset1:   asset1,
		},
	}, nil
}

// Pools returns all concentrated pools in existence.
func (q Querier) Pools(
	ctx sdk.Context,
	req clquery.PoolsRequest,
) (*clquery.PoolsResponse, error) {
	anys, pageRes, err := q.Keeper.GetSerializedPools(ctx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &clquery.PoolsResponse{
		Pools:      anys,
		Pagination: pageRes,
	}, nil
}

// Params returns module params
func (q Querier) Params(ctx sdk.Context, req clquery.ParamsRequest) (*clquery.ParamsResponse, error) {
	return &clquery.ParamsResponse{Params: q.Keeper.GetParams(ctx)}, nil
}

// LiquidityPerTickRange returns the amount of liquidity per every tick range
// existing within the given pool. The amounts are returned as a slice of ranges with their liquidity depths.
func (q Querier) LiquidityPerTickRange(ctx sdk.Context, req clquery.LiquidityPerTickRangeRequest) (*clquery.LiquidityPerTickRangeResponse, error) {
	liquidity, err := q.Keeper.GetTickLiquidityForFullRange(
		ctx,
		req.PoolId,
	)
	if err != nil {
		return nil, err
	}

	return &clquery.LiquidityPerTickRangeResponse{Liquidity: liquidity}, nil
}

// LiquidityNetInDirection returns an array of LiquidityDepthWithRange, which contains the range(lower tick and upper tick) and the liquidity amount in the range.
func (q Querier) LiquidityNetInDirection(ctx sdk.Context, req clquery.LiquidityNetInDirectionRequest) (*clquery.LiquidityNetInDirectionResponse, error) {
	if req.TokenIn == "" {
		return nil, status.Error(codes.InvalidArgument, "tokenIn is empty")
	}

	var startTick sdk.Int
	if !req.UseCurTick {
		startTick = sdk.NewInt(req.StartTick)
	}

	var boundTick sdk.Int
	if !req.UseNoBound {
		boundTick = sdk.NewInt(req.BoundTick)
	}

	liquidityDepths, err := q.Keeper.GetTickLiquidityNetInDirection(
		ctx,
		req.PoolId,
		req.TokenIn,
		startTick,
		boundTick,
	)
	if err != nil {
		return nil, err
	}

	pool, err := q.Keeper.GetPoolFromPoolIdAndConvertToConcentrated(ctx, req.PoolId)
	if err != nil {
		return nil, err
	}

	return &clquery.LiquidityNetInDirectionResponse{LiquidityDepths: liquidityDepths, CurrentLiquidity: pool.GetLiquidity(), CurrentTick: pool.GetCurrentTick().Int64()}, nil
}

func (q Querier) ClaimableFees(ctx sdk.Context, req clquery.ClaimableFeesRequest) (*clquery.ClaimableFeesResponse, error) {
	claimableFees, err := q.Keeper.GetClaimableFees(ctx, req.PositionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &clquery.ClaimableFeesResponse{
		ClaimableFees: claimableFees,
	}, nil
}

func (q Querier) ClaimableIncentives(ctx sdk.Context, req clquery.ClaimableIncentivesRequest) (*clquery.ClaimableIncentivesResponse, error) {
	claimableIncentives, forfeitedIncentives, err := q.Keeper.GetClaimableIncentives(ctx, req.PositionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &clquery.ClaimableIncentivesResponse{
		ClaimableIncentives: claimableIncentives,
		ForfeitedIncentives: forfeitedIncentives,
	}, nil
}
