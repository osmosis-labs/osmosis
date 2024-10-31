package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	incentivetypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/pool-incentives keeper providing gRPC
// method handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// GaugeIds takes provided gauge request and returns the respective internally incentivized gaugeIDs.
// If internally incentivized for a given pool id is not found, returns an error.
func (q Querier) GaugeIds(ctx context.Context, req *types.QueryGaugeIdsRequest) (*types.QueryGaugeIdsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	distrInfo := q.Keeper.GetDistrInfo(sdkCtx)
	totalWeightDec := distrInfo.TotalWeight.ToLegacyDec()
	incentivePercentage := osmomath.NewDec(0)
	percentMultiplier := osmomath.NewInt(100)

	pool, err := q.Keeper.poolmanagerKeeper.GetPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	isConcentratedPool := pool.GetType() == poolmanagertypes.Concentrated
	if isConcentratedPool {
		incentiveEpochDuration := q.Keeper.incentivesKeeper.GetEpochInfo(sdkCtx).Duration
		gaugeId, err := q.Keeper.GetPoolGaugeId(sdkCtx, req.PoolId, incentiveEpochDuration)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		for _, record := range distrInfo.Records {
			if record.GaugeId == gaugeId {
				// Pool incentive % = (gauge_id_weight / sum_of_all_pool_gauge_weight) * 100
				incentivePercentage = record.Weight.ToLegacyDec().Quo(totalWeightDec).MulInt(percentMultiplier)
			}
		}

		return &types.QueryGaugeIdsResponse{
			GaugeIdsWithDuration: []*types.QueryGaugeIdsResponse_GaugeIdWithDuration{
				{
					GaugeId:                  gaugeId,
					Duration:                 incentiveEpochDuration,
					GaugeIncentivePercentage: incentivePercentage.String(),
				},
			},
		}, nil
	}

	lockableDurations := q.Keeper.GetLockableDurations(sdkCtx)
	gaugeIdsWithDuration := make([]*types.QueryGaugeIdsResponse_GaugeIdWithDuration, len(lockableDurations))

	for i, duration := range lockableDurations {
		gaugeId, err := q.Keeper.GetPoolGaugeId(sdkCtx, req.PoolId, duration)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		for _, record := range distrInfo.Records {
			if record.GaugeId == gaugeId {
				// Pool incentive % = (gauge_id_weight / sum_of_all_pool_gauge_weight) * 100
				incentivePercentage = record.Weight.ToLegacyDec().Quo(totalWeightDec).MulInt(percentMultiplier)
			}
		}

		gaugeIdsWithDuration[i] = &types.QueryGaugeIdsResponse_GaugeIdWithDuration{
			GaugeId:                  gaugeId,
			Duration:                 duration,
			GaugeIncentivePercentage: incentivePercentage.String(),
		}
	}
	return &types.QueryGaugeIdsResponse{GaugeIdsWithDuration: gaugeIdsWithDuration}, nil
}

// DistrInfo returns gauges receiving pool rewards and their respective weights.
func (q Querier) DistrInfo(ctx context.Context, _ *types.QueryDistrInfoRequest) (*types.QueryDistrInfoResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &types.QueryDistrInfoResponse{DistrInfo: q.Keeper.GetDistrInfo(sdkCtx)}, nil
}

// Params return pool-incentives module params.
func (q Querier) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &types.QueryParamsResponse{Params: q.Keeper.GetParams(sdkCtx)}, nil
}

// LockableDurations returns the lock durations that are incentivized through pool-incentives.
func (q Querier) LockableDurations(ctx context.Context, _ *types.QueryLockableDurationsRequest) (*types.QueryLockableDurationsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &types.QueryLockableDurationsResponse{LockableDurations: q.Keeper.GetLockableDurations(sdkCtx)}, nil
}

// IncentivizedPools iterates over all internally incentivized gauges and returns their corresponding pools.
func (q Querier) IncentivizedPools(ctx context.Context, _ *types.QueryIncentivizedPoolsRequest) (*types.QueryIncentivizedPoolsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// We use lockable durations for byDuration gauges.
	lockableDurations := q.Keeper.GetLockableDurations(sdkCtx)
	distrInfo := q.Keeper.GetDistrInfo(sdkCtx)

	// We use epoch duration for CL noLock gauges.
	epochDuration := q.incentivesKeeper.GetEpochInfo(sdkCtx).Duration

	// While there are exceptions, typically the number of incentivizedPools
	// equals to the number of incentivized gauges / number of lockable durations.
	incentivizedPools := make([]types.IncentivizedPool, 0, len(distrInfo.Records)/len(lockableDurations))

	// Loop over the distribution records and fill in the incentivized pools struct.
	for _, record := range distrInfo.Records {
		// Skip community pool gauge
		if record.GaugeId == 0 {
			continue
		}
		gauge, err := q.incentivesKeeper.GetGaugeByID(sdkCtx, record.GaugeId)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		if gauge.DistributeTo.LockQueryType == lockuptypes.ByGroup {
			group, err := q.Keeper.incentivesKeeper.GetGroupByGaugeID(sdkCtx, record.GaugeId)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			groupGauge, err := q.Keeper.incentivesKeeper.GetGaugeByID(sdkCtx, group.GroupGaugeId)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			if !groupGauge.IsPerpetual {
				// if the group is not perpetual, it is an externally incentivized gauge so we skip it
				continue
			}
			poolIds, durations, err := q.Keeper.incentivesKeeper.GetPoolIdsAndDurationsFromGaugeRecords(sdkCtx, group.InternalGaugeInfo.GaugeRecords)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			for i, poolId := range poolIds {
				incentivizedPool := types.IncentivizedPool{
					PoolId:           poolId,
					LockableDuration: durations[i],
					GaugeId:          record.GaugeId,
				}

				incentivizedPools = append(incentivizedPools, incentivizedPool)
			}
		} else if gauge.DistributeTo.LockQueryType == lockuptypes.NoLock {
			poolId, err := q.Keeper.GetPoolIdFromGaugeId(sdkCtx, record.GaugeId, epochDuration)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			incentivizedPool := types.IncentivizedPool{
				PoolId:           poolId,
				LockableDuration: epochDuration,
				GaugeId:          record.GaugeId,
			}

			incentivizedPools = append(incentivizedPools, incentivizedPool)
		} else if gauge.DistributeTo.LockQueryType == lockuptypes.ByDuration {
			gauge, err := q.incentivesKeeper.GetGaugeByID(sdkCtx, record.GaugeId)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			// Ensure the gauge's duration matches one of the lockable durations.
			matchFound := false
			for _, duration := range lockableDurations {
				if gauge.DistributeTo.Duration == duration {
					matchFound = true
					break
				}
			}
			if !matchFound {
				return nil, types.IncentiveRecordContainsNonLockableDurationError{GaugeId: gauge.Id, Duration: gauge.DistributeTo.Duration, LockableDurations: lockableDurations}
			}
			poolId, err := q.Keeper.GetPoolIdFromGaugeId(sdkCtx, record.GaugeId, gauge.DistributeTo.Duration)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			incentivizedPool := types.IncentivizedPool{
				PoolId:           poolId,
				LockableDuration: gauge.DistributeTo.Duration,
				GaugeId:          record.GaugeId,
			}

			incentivizedPools = append(incentivizedPools, incentivizedPool)
		} else {
			return nil, status.Error(codes.Internal, "unknown lock query type")
		}
	}

	return &types.QueryIncentivizedPoolsResponse{
		IncentivizedPools: incentivizedPools,
	}, nil
}

// ExternalIncentiveGauges iterates over all gauges and returns gauges externally incentivized by excluding default (internal) gauges.
func (q Querier) ExternalIncentiveGauges(ctx context.Context, req *types.QueryExternalIncentiveGaugesRequest) (*types.QueryExternalIncentiveGaugesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(q.Keeper.storeKey)
	prefixStore := prefix.NewStore(store, []byte("pool-incentives/"))

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	// map true to default gauges created with pool
	poolGaugeIds := make(map[uint64]bool)
	for ; iterator.Valid(); iterator.Next() {
		poolGaugeIds[sdk.BigEndianToUint64(iterator.Value())] = true
	}

	// iterate over all gauges, exclude default created gauges, leaving externally incentivized gauges
	allGauges := q.Keeper.GetAllGauges(sdkCtx)
	gauges := []incentivetypes.Gauge{}
	for _, gauge := range allGauges {
		if _, ok := poolGaugeIds[gauge.Id]; !ok {
			gauges = append(gauges, gauge)
		}
	}

	return &types.QueryExternalIncentiveGaugesResponse{Data: gauges}, nil
}
