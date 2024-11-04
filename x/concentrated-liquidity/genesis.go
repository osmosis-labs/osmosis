package concentrated_liquidity

import (
	"errors"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	types "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types/genesis"
)

// InitGenesis initializes the concentrated-liquidity module with the provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState genesis.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.SetNextPositionId(ctx, genState.NextPositionId)
	k.SetNextIncentiveRecordId(ctx, genState.NextIncentiveRecordId)
	// Initialize pools
	totalLiquidity := sdk.Coins{}
	var unpacker codectypes.AnyUnpacker = k.cdc
	seenPoolIds := map[uint64]struct{}{}
	for _, poolData := range genState.PoolData {
		var pool types.ConcentratedPoolExtension
		err := unpacker.UnpackAny(poolData.Pool, &pool)
		if err != nil {
			panic(err)
		}
		err = k.setPool(ctx, pool)
		if err != nil {
			panic(err)
		}

		poolId := pool.GetId()
		poolTicks := poolData.Ticks
		for _, tick := range poolTicks {
			k.SetTickInfo(ctx, poolId, tick.TickIndex, &tick.Info)
		}
		seenPoolIds[poolId] = struct{}{}

		// add pools liquidity to total liquidity
		poolLiquidity, err := k.GetTotalPoolLiquidity(ctx, poolId)
		if err != nil {
			panic(err)
		}
		totalLiquidity = totalLiquidity.Add(poolLiquidity...)

		// set up spread reward accumulators
		store := ctx.KVStore(k.storeKey)
		err = accum.MakeAccumulatorWithValueAndShare(store, poolData.SpreadRewardAccumulator.Name, poolData.SpreadRewardAccumulator.AccumContent.AccumValue, poolData.SpreadRewardAccumulator.AccumContent.TotalShares)
		if err != nil {
			panic(err)
		}

		// set up incentive accumulators
		for _, incentiveAccum := range poolData.IncentivesAccumulators {
			err = accum.MakeAccumulatorWithValueAndShare(store, incentiveAccum.GetName(), incentiveAccum.AccumContent.AccumValue, incentiveAccum.AccumContent.TotalShares)
			if err != nil {
				panic(err)
			}
		}

		// set incentive records
		err = k.setMultipleIncentiveRecords(ctx, poolData.IncentiveRecords)
		if err != nil {
			panic(err)
		}
	}

	// set positions for pool
	for _, positionWrapper := range genState.PositionData {
		if _, ok := seenPoolIds[positionWrapper.Position.PoolId]; !ok {
			panic(fmt.Sprintf("found position with pool id (%d) but there is no pool with such id that exists", positionWrapper.Position.PoolId))
		}

		err := k.SetPosition(ctx, positionWrapper.Position.PoolId, sdk.MustAccAddressFromBech32(positionWrapper.Position.Address), positionWrapper.Position.LowerTick, positionWrapper.Position.UpperTick, positionWrapper.Position.JoinTime, positionWrapper.Position.Liquidity, positionWrapper.Position.PositionId, positionWrapper.LockId)
		if err != nil {
			panic(err)
		}

		// set individual spread reward accumulator state position
		spreadRewardAccumObject, err := k.GetSpreadRewardAccumulator(ctx, positionWrapper.Position.PoolId)
		if err != nil {
			panic(err)
		}
		spreadRewardPositionKey := types.KeySpreadRewardPositionAccumulator(positionWrapper.Position.PositionId)

		k.initOrUpdateAccumPosition(ctx, spreadRewardAccumObject, positionWrapper.SpreadRewardAccumRecord.AccumValuePerShare, spreadRewardPositionKey, positionWrapper.SpreadRewardAccumRecord.NumShares, positionWrapper.SpreadRewardAccumRecord.UnclaimedRewardsTotal, positionWrapper.SpreadRewardAccumRecord.Options)

		positionName := string(types.KeyPositionId(positionWrapper.Position.PositionId))
		uptimeAccumulators, err := k.GetUptimeAccumulators(ctx, positionWrapper.Position.PoolId)
		if err != nil {
			panic(err)
		}

		for uptimeIndex, uptimeRecord := range positionWrapper.UptimeAccumRecords {
			k.initOrUpdateAccumPosition(ctx, uptimeAccumulators[uptimeIndex], uptimeRecord.AccumValuePerShare, positionName, uptimeRecord.NumShares, uptimeRecord.UnclaimedRewardsTotal, uptimeRecord.Options)
		}
	}

	// set total liquidity
	k.setTotalLiquidity(ctx, totalLiquidity)

	k.SetIncentivePoolIDMigrationThreshold(ctx, genState.IncentivesAccumulatorPoolIdMigrationThreshold)
	k.SetSpreadFactorPoolIDMigrationThreshold(ctx, genState.SpreadFactorPoolIdMigrationThreshold)
}

// ExportGenesis returns the concentrated-liquidity module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *genesis.GenesisState {
	pools, err := k.GetPools(ctx)
	if err != nil {
		panic(err)
	}

	poolData := make([]genesis.PoolData, 0, len(pools))

	for _, poolI := range pools {
		poolI := poolI
		any, err := codectypes.NewAnyWithValue(poolI)
		if err != nil {
			panic(err)
		}
		anyCopy := *any

		ticks, err := k.GetAllInitializedTicksForPool(ctx, poolI.GetId())
		if err != nil {
			panic(err)
		}
		accumObject, err := k.GetSpreadRewardAccumulator(ctx, poolI.GetId())
		if err != nil {
			panic(err)
		}

		totalShares := accumObject.GetTotalShares()

		spreadRewardAccumObject := genesis.AccumObject{
			Name: types.KeySpreadRewardPoolAccumulator(poolI.GetId()),
			AccumContent: &accum.AccumulatorContent{
				AccumValue:  accumObject.GetValue(),
				TotalShares: totalShares,
			},
		}

		poolId := poolI.GetId()
		incentiveRecordsForPool, err := k.GetAllIncentiveRecordsForPool(ctx, poolId)
		if err != nil {
			panic(err)
		}

		incentivesAccum, err := k.GetUptimeAccumulators(ctx, poolId)
		if err != nil {
			panic(err)
		}

		incentivesAccumObject := make([]genesis.AccumObject, len(incentivesAccum))
		for i, incentiveAccum := range incentivesAccum {
			incentiveAccumTotalShares := incentiveAccum.GetTotalShares()

			genesisAccum := genesis.AccumObject{
				Name: incentiveAccum.GetName(),
				AccumContent: &accum.AccumulatorContent{
					AccumValue:  incentiveAccum.GetValue(),
					TotalShares: incentiveAccumTotalShares,
				},
			}
			incentivesAccumObject[i] = genesisAccum
		}

		poolData = append(poolData, genesis.PoolData{
			Pool:                    &anyCopy,
			Ticks:                   ticks,
			SpreadRewardAccumulator: spreadRewardAccumObject,
			IncentivesAccumulators:  incentivesAccumObject,
			IncentiveRecords:        incentiveRecordsForPool,
		})
	}

	positions, err := k.getAllPositions(ctx)
	if err != nil {
		panic(err)
	}

	positionData := make([]genesis.PositionData, 0, len(positions))
	for _, position := range positions {
		position, err := k.GetPosition(ctx, position.PositionId)
		if err != nil {
			panic(err)
		}

		lockId, err := k.GetLockIdFromPositionId(ctx, position.PositionId)
		if err != nil {
			if !errors.Is(err, types.PositionIdToLockNotFoundError{PositionId: position.PositionId}) {
				panic(err)
			}
		}

		// Retrieve spread reward accumulator state for position
		spreadRewardPositionKey := types.KeySpreadRewardPositionAccumulator(position.PositionId)
		spreadRewardAccumObject, err := k.GetSpreadRewardAccumulator(ctx, position.PoolId)
		if err != nil {
			panic(err)
		}
		spreadRewardAccumPositionRecord, err := spreadRewardAccumObject.GetPosition(spreadRewardPositionKey)
		if err != nil {
			panic(err)
		}

		// Retrieve uptime incentive accumulator state for position
		positionName := string(types.KeyPositionId(position.PositionId))
		uptimeAccumulators, err := k.GetUptimeAccumulators(ctx, position.PoolId)
		if err != nil {
			panic(err)
		}

		uptimeAccumObject := make([]accum.Record, len(uptimeAccumulators))
		for uptimeIndex := range types.SupportedUptimes {
			accumRecord, err := uptimeAccumulators[uptimeIndex].GetPosition(positionName)
			if err != nil {
				panic(err)
			}

			uptimeAccumObject[uptimeIndex] = accumRecord
		}

		positionData = append(positionData, genesis.PositionData{
			LockId:                  lockId,
			Position:                &position,
			SpreadRewardAccumRecord: spreadRewardAccumPositionRecord,
			UptimeAccumRecords:      uptimeAccumObject,
		})
	}

	// Get the incentive pool ID migration threshold
	incentivesAccumulatorPoolIDMigrationThreshold, err := k.GetIncentivePoolIDMigrationThreshold(ctx)
	if err != nil {
		panic(err)
	}

	// Get the spread factor pool ID migration threshold
	spreadFactorPoolIdMigrationThreshold, err := k.GetSpreadFactorPoolIDMigrationThreshold(ctx)
	if err != nil {
		panic(err)
	}

	return &genesis.GenesisState{
		Params:                k.GetParams(ctx),
		PoolData:              poolData,
		PositionData:          positionData,
		NextPositionId:        k.GetNextPositionId(ctx),
		NextIncentiveRecordId: k.GetNextIncentiveRecordId(ctx),
		IncentivesAccumulatorPoolIdMigrationThreshold: incentivesAccumulatorPoolIDMigrationThreshold,
		SpreadFactorPoolIdMigrationThreshold:          spreadFactorPoolIdMigrationThreshold,
	}
}

// initOrUpdateAccumPosition creates a new position or override an existing position
// at accumulator's current value with a specific number of shares and unclaimed rewards
func (k Keeper) initOrUpdateAccumPosition(ctx sdk.Context, accumumulator *accum.AccumulatorObject, accumulatorValuePerShare sdk.DecCoins, index string, numShareUnits osmomath.Dec, unclaimedRewardsTotal sdk.DecCoins, options *accum.Options) {
	position := accum.Record{
		NumShares:             numShareUnits,
		AccumValuePerShare:    accumulatorValuePerShare,
		UnclaimedRewardsTotal: unclaimedRewardsTotal,
		Options:               options,
	}

	osmoutils.MustSet(ctx.KVStore(k.storeKey), accum.FormatPositionPrefixKey(accumumulator.GetName(), index), &position)
}
