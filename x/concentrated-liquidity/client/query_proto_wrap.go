package client

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity"
	clquery "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/client/queryproto"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/model"
)

// Querier defines a wrapper around the x/concentrated-liquidity keeper providing gRPC method
// handlers.
type Querier struct {
	cl.Keeper
}

func NewQuerier(k cl.Keeper) Querier {
	return Querier{Keeper: k}
}

// UserPositions returns positions of a specified address. Each position is broken down by:
// - the position itself
// - the underlying assets
// - the claimable fees
// - the claimable incentives
// - the incentives that would be forfeited if the position was closed now
func (q Querier) UserPositions(ctx sdk.Context, req clquery.UserPositionsRequest) (*clquery.UserPositionsResponse, error) {
	sdkAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	fullPositions, pageRes, err := q.Keeper.GetUserPositionsSerialized(ctx, sdkAddr, req.PoolId, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &clquery.UserPositionsResponse{
		Positions:  fullPositions,
		Pagination: pageRes,
	}, nil
}

// PositionById returns a position with the specified id. The position is broken down by:
// - the position itself
// - the underlying assets
// - the claimable fees
// - the claimable incentives
// - the incentives that would be forfeited if the position was closed now
func (q Querier) PositionById(ctx sdk.Context, req clquery.PositionByIdRequest) (*clquery.PositionByIdResponse, error) {
	position, err := q.Keeper.GetPosition(ctx, req.PositionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	positionPool, err := q.Keeper.GetConcentratedPoolById(ctx, position.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	asset0, asset1, err := cl.CalculateUnderlyingAssetsFromPosition(ctx, position, positionPool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	claimableSpreadRewards, err := q.Keeper.GetClaimableSpreadRewards(ctx, position.PositionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	claimableIncentives, forfeitedIncentives, err := q.Keeper.GetClaimableIncentives(ctx, position.PositionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &clquery.PositionByIdResponse{
		Position: model.FullPositionBreakdown{
			Position:               position,
			Asset0:                 asset0,
			Asset1:                 asset1,
			ClaimableSpreadRewards: claimableSpreadRewards,
			ClaimableIncentives:    claimableIncentives,
			ForfeitedIncentives:    forfeitedIncentives,
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

	pool, err := q.Keeper.GetConcentratedPoolById(ctx, req.PoolId)
	if err != nil {
		return nil, err
	}

	return &clquery.LiquidityNetInDirectionResponse{LiquidityDepths: liquidityDepths, CurrentLiquidity: pool.GetLiquidity(), CurrentTick: pool.GetCurrentTick()}, nil
}

func (q Querier) ClaimableSpreadRewards(ctx sdk.Context, req clquery.ClaimableSpreadRewardsRequest) (*clquery.ClaimableSpreadRewardsResponse, error) {
	ClaimableSpreadRewards, err := q.Keeper.GetClaimableSpreadRewards(ctx, req.PositionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &clquery.ClaimableSpreadRewardsResponse{
		ClaimableSpreadRewards: ClaimableSpreadRewards,
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

// PoolAccumulatorRewards returns pool accumulator rewards.
// It includes global spread reward growth and global uptime growth accumulator values.
func (q Querier) PoolAccumulatorRewards(ctx sdk.Context, req clquery.PoolAccumulatorRewardsRequest) (*clquery.PoolAccumulatorRewardsResponse, error) {
	if req.PoolId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pool id is zero")
	}

	// We utilize a cache context here as we need to update the global uptime accumulators but
	// we don't want to persist the changes to the store.
	cacheCtx, _ := ctx.CacheContext()

	// Sync global uptime accumulators to ensure the uptime tracker init values are up to date.
	err := q.Keeper.UpdatePoolUptimeAccumulatorsToNow(cacheCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	spreadRewardsAcc, err := q.Keeper.GetSpreadRewardAccumulator(cacheCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	uptimeAccValues, err := q.Keeper.GetUptimeAccumulatorValues(cacheCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	uptimeGrowthTrackers := make([]model.UptimeTracker, 0, len(uptimeAccValues))
	for _, uptimeTrackerValue := range uptimeAccValues {
		uptimeGrowthTrackers = append(uptimeGrowthTrackers, model.UptimeTracker{UptimeGrowthOutside: uptimeTrackerValue})
	}

	return &clquery.PoolAccumulatorRewardsResponse{
		SpreadRewardGrowthGlobal: spreadRewardsAcc.GetValue(),
		UptimeGrowthGlobal:       uptimeGrowthTrackers,
	}, nil
}

func (q Querier) IncentiveRecords(ctx sdk.Context, req clquery.IncentiveRecordsRequest) (*clquery.IncentiveRecordsResponse, error) {
	anys, pageRes, err := q.Keeper.GetIncentiveRecordSerialized(ctx, req.PoolId, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &clquery.IncentiveRecordsResponse{
		IncentiveRecords: anys,
		Pagination:       pageRes,
	}, nil
}

// TickAccumulatorTrackers returns tick accumulator trackers.
// It includes spread reward growth in the opposite direction of last traversal and uptime tracker values.
func (q Querier) TickAccumulatorTrackers(ctx sdk.Context, req clquery.TickAccumulatorTrackersRequest) (*clquery.TickAccumulatorTrackersResponse, error) {
	if req.PoolId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pool id is zero")
	}

	cacheCtx, _ := ctx.CacheContext()
	tickInfo, err := q.Keeper.GetTickInfo(cacheCtx, req.PoolId, req.TickIndex)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &clquery.TickAccumulatorTrackersResponse{
		SpreadRewardGrowthOppositeDirectionOfLastTraversal: tickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal,
		UptimeTrackers: tickInfo.UptimeTrackers.List,
	}, nil
}

// CFMMPoolIdLinkFromConcentratedPoolId queries the cfmm pool id linked to a concentrated pool id.
func (q Querier) CFMMPoolIdLinkFromConcentratedPoolId(ctx sdk.Context, req clquery.CFMMPoolIdLinkFromConcentratedPoolIdRequest) (*clquery.CFMMPoolIdLinkFromConcentratedPoolIdResponse, error) {
	if req.ConcentratedPoolId == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid cfmm pool id")
	}
	cfmmPoolId, err := q.Keeper.GetLinkedBalancerPoolID(ctx, req.ConcentratedPoolId)
	if err != nil {
		return nil, err
	}

	return &clquery.CFMMPoolIdLinkFromConcentratedPoolIdResponse{
		CfmmPoolId: cfmmPoolId,
	}, nil
}

// UserUnbodingPositions returns all the unbonding concentrated liquidity positions along with their respective period lock.
func (q Querier) UserUnbondingPositions(ctx sdk.Context, req clquery.UserUnbondingPositionsRequest) (*clquery.UserUnbondingPositionsResponse, error) {
	sdkAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	cfmmPoolId, err := q.Keeper.GetUserUnbondingPositions(ctx, sdkAddr)
	if err != nil {
		return nil, err
	}

	return &clquery.UserUnbondingPositionsResponse{
		PositionsWithPeriodLock: cfmmPoolId,
	}, nil
}

// GetTotalLiquidity returns the total liquidity across all concentrated liquidity pools.
func (q Querier) GetTotalLiquidity(ctx sdk.Context, req clquery.GetTotalLiquidityRequest) (*clquery.GetTotalLiquidityResponse, error) {
	totalLiquidity, err := q.Keeper.GetTotalLiquidity(ctx)
	if err != nil {
		return nil, err
	}

	return &clquery.GetTotalLiquidityResponse{
		TotalLiquidity: totalLiquidity,
	}, nil
}
