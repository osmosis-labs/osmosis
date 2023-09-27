package keeper

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	incentivetypes "github.com/osmosis-labs/osmosis/v19/x/incentives/types"
	"github.com/osmosis-labs/osmosis/v19/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
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

	lockableDurations := q.Keeper.GetLockableDurations(sdkCtx)
	distrInfo := q.Keeper.GetDistrInfo(sdkCtx)

	// Add epoch duration to lockable durations if not already present.
	// This is to ensure that concentrated gauges (which run on epoch time) are
	// always included in the query, even if the epoch duration changes in the future.
	epochDuration := q.incentivesKeeper.GetEpochInfo(sdkCtx).Duration
	epochAlreadyLockable := false
	for _, lockableDuration := range lockableDurations {
		if lockableDuration == epochDuration {
			epochAlreadyLockable = true
			break
		}
	}

	// Ensure that we only add epoch duration if it does not already exist as a lockable duration.
	if !epochAlreadyLockable {
		lockableDurations = append(lockableDurations, epochDuration)
	}

	// While there are exceptions, typically the number of incentivizedPools
	// equals to the number of incentivized gauges / number of lockable durations.
	incentivizedPools := make([]types.IncentivizedPool, 0, len(distrInfo.Records)/len(lockableDurations))
	incentivizedPoolIDs := make(map[uint64]time.Duration)

	// Loop over the distribution records and fill in the incentivized pools struct.
	for _, record := range distrInfo.Records {
		for _, lockableDuration := range lockableDurations {
			poolId, err := q.Keeper.GetPoolIdFromGaugeId(sdkCtx, record.GaugeId, lockableDuration)
			if err == nil {
				incentivizedPool := types.IncentivizedPool{
					PoolId:           poolId,
					LockableDuration: lockableDuration,
					GaugeId:          record.GaugeId,
				}

				incentivizedPools = append(incentivizedPools, incentivizedPool)
				incentivizedPoolIDs[poolId] = lockableDuration
			}
		}
	}

	// Only run the following if the above loop determined there were incentivized pools.
	if len(incentivizedPoolIDs) > 0 {
		// Retrieve the migration records between balancer pools and concentrated liquidity pools.
		// This comes from the superfluid keeper, since superfluid is the only pool incentives connected
		// module that has access to the gamm modules store.
		migrationRecords, err := q.gammKeeper.GetAllMigrationInfo(sdkCtx)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Iterate over all migration records.
		for _, record := range migrationRecords.BalancerToConcentratedPoolLinks {
			// If the cl pool is not in the list of incentivized pools, skip it.
			lockableDuration, incentivized := incentivizedPoolIDs[record.ClPoolId]
			if !incentivized {
				continue
			}

			// Add the indirectly incentivized balancer pools to the list of incentivized pools.
			gaugeId, err := q.Keeper.GetPoolGaugeId(sdkCtx, record.BalancerPoolId, lockableDuration)
			if err == nil {
				incentivizedPool := types.IncentivizedPool{
					PoolId:           record.BalancerPoolId,
					LockableDuration: lockableDuration,
					GaugeId:          gaugeId,
				}

				incentivizedPools = append(incentivizedPools, incentivizedPool)
			}
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
