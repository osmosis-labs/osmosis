package concentrated_liquidity

import (
	"bytes"
	"errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	types "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types/genesis"
)

// InitGenesis initializes the concentrated-liquidity module with the provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState genesis.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.SetNextPositionId(ctx, genState.NextPositionId)
	k.SetNextIncentiveRecordId(ctx, genState.NextIncentiveRecordId)
	// Initialize pools
	var unpacker codectypes.AnyUnpacker = k.cdc
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

		// set positions for pool
		for _, positionWrapper := range poolData.PositionData {
			err := k.SetPosition(ctx, poolId, sdk.MustAccAddressFromBech32(positionWrapper.Position.Address), positionWrapper.Position.LowerTick, positionWrapper.Position.UpperTick, positionWrapper.Position.JoinTime, positionWrapper.Position.Liquidity, positionWrapper.Position.PositionId, positionWrapper.LockId)
			if err != nil {
				panic(err)
			}

			// set individual spread reward accumulator state position
			spreadRewardAccumObject, err := k.GetSpreadRewardAccumulator(ctx, poolId)
			if err != nil {
				panic(err)
			}
			spreadRewardPositionKey := types.KeySpreadRewardPositionAccumulator(positionWrapper.Position.PositionId)

			k.initOrUpdateAccumPosition(ctx, spreadRewardAccumObject, positionWrapper.SpreadRewardAccumRecord.AccumValuePerShare, spreadRewardPositionKey, positionWrapper.SpreadRewardAccumRecord.NumShares, positionWrapper.SpreadRewardAccumRecord.UnclaimedRewardsTotal, positionWrapper.SpreadRewardAccumRecord.Options)

			positionName := string(types.KeyPositionId(positionWrapper.Position.PositionId))
			uptimeAccumulators, err := k.GetUptimeAccumulators(ctx, poolId)
			if err != nil {
				panic(err)
			}

			for uptimeIndex, uptimeRecord := range positionWrapper.UptimeAccumRecords {
				k.initOrUpdateAccumPosition(ctx, uptimeAccumulators[uptimeIndex], uptimeRecord.AccumValuePerShare, positionName, uptimeRecord.NumShares, uptimeRecord.UnclaimedRewardsTotal, uptimeRecord.Options)
			}
		}
	}
}

// ExportGenesis returns the concentrated-liquidity module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *genesis.GenesisState {
	positions, err := k.getAllPositions(ctx)
	if err != nil {
		panic(err)
	}

	positionDataMap := map[uint64][]genesis.PositionData{}
	for _, position := range positions {
		position, err := k.GetPosition(ctx, position.PositionId)
		if err != nil {
			panic(err)
		}
		positionWithoutPoolId := genesis.PositionWithoutPoolId{}
		positionWithoutPoolId.Address = position.Address
		positionWithoutPoolId.JoinTime = position.JoinTime
		positionWithoutPoolId.Liquidity = position.Liquidity
		positionWithoutPoolId.LowerTick = position.LowerTick
		positionWithoutPoolId.UpperTick = position.UpperTick
		positionWithoutPoolId.PositionId = position.PositionId

		lockId, err := k.GetLockIdFromPositionId(ctx, position.PositionId)
		if err != nil {
			if errors.Is(err, types.PositionIdToLockNotFoundError{PositionId: position.PositionId}) {
				lockId = 0
			} else {
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

		if positionDataMap[position.PoolId] == nil {
			positionDataMap[position.PoolId] = make([]genesis.PositionData, 0)
		}

		positionDataMap[position.PoolId] = append(positionDataMap[position.PoolId], genesis.PositionData{
			LockId:                  lockId,
			Position:                &positionWithoutPoolId,
			SpreadRewardAccumRecord: spreadRewardAccumPositionRecord,
			UptimeAccumRecords:      uptimeAccumObject,
		})
	}

	pools, err := k.GetPools(ctx)
	if err != nil {
		panic(err)
	}

	poolData := make([]genesis.GenesisPoolData, 0, len(pools))

	for _, poolI := range pools {
		poolI := poolI
		any, err := codectypes.NewAnyWithValue(poolI)
		if err != nil {
			panic(err)
		}
		anyCopy := *any

		ticks, err := k.GetAllInitializedTicksForPoolWithoutPoolId(ctx, poolI.GetId())
		if err != nil {
			panic(err)
		}
		accumObject, err := k.GetSpreadRewardAccumulator(ctx, poolI.GetId())
		if err != nil {
			panic(err)
		}

		totalShares, err := accumObject.GetTotalShares()
		if err != nil {
			panic(err)
		}

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
			incentiveAccumTotalShares, err := incentiveAccum.GetTotalShares()
			if err != nil {
				panic(err)
			}
			genesisAccum := genesis.AccumObject{
				Name: incentiveAccum.GetName(),
				AccumContent: &accum.AccumulatorContent{
					AccumValue:  incentiveAccum.GetValue(),
					TotalShares: incentiveAccumTotalShares,
				},
			}
			incentivesAccumObject[i] = genesisAccum
		}

		positionData := make([]genesis.PositionData, 0)
		if len(positionDataMap[poolId]) > 0 {
			positionData = positionDataMap[poolId]
		}

		poolData = append(poolData, genesis.GenesisPoolData{
			Pool:                    &anyCopy,
			PositionData:            positionData,
			Ticks:                   ticks,
			SpreadRewardAccumulator: spreadRewardAccumObject,
			IncentivesAccumulators:  incentivesAccumObject,
			IncentiveRecords:        incentiveRecordsForPool,
		})
	}

	return &genesis.GenesisState{
		Params:                k.GetParams(ctx),
		PoolData:              poolData,
		NextPositionId:        k.GetNextPositionId(ctx),
		NextIncentiveRecordId: k.GetNextIncentiveRecordId(ctx),
	}
}

// initOrUpdateAccumPosition creates a new position or override an existing position
// at accumulator's current value with a specific number of shares and unclaimed rewards
func (k Keeper) initOrUpdateAccumPosition(ctx sdk.Context, accumumulator accum.AccumulatorObject, accumulatorValuePerShare sdk.DecCoins, index string, numShareUnits sdk.Dec, unclaimedRewardsTotal sdk.DecCoins, options *accum.Options) {
	position := accum.Record{
		NumShares:             numShareUnits,
		AccumValuePerShare:    accumulatorValuePerShare,
		UnclaimedRewardsTotal: unclaimedRewardsTotal,
		Options:               options,
	}

	osmoutils.MustSet(ctx.KVStore(k.storeKey), accum.FormatPositionPrefixKey(accumumulator.GetName(), index), &position)
}

// GetAllInitializedTicksForPoolWithoutPoolId returns all ticks in FullTick struct, without the pool id.
func (k Keeper) GetAllInitializedTicksForPoolWithoutPoolId(ctx sdk.Context, poolId uint64) ([]genesis.FullTick, error) {
	return osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), types.KeyTickPrefixByPoolId(poolId), ParseFullTickFromBytes)
}

// ParseFullTickFromBytes takes key and value byte slices and attempts to parse
// them into a FullTick struct. If the key or value is not valid, an appropriate
// error is returned. The function expects the key to have three components
// 1. The tick prefix (1 byte)
// 2. The pool id (8 bytes)
// 3. The tick index (1 byte for sign + 8 bytes for unsigned integer)
//
// The function returns a FullTick struct containing the tick index, and
// tick information.
//
// Parameters:
// - key ([]byte): A byte slice representing the key.
// - value ([]byte): A byte slice representing the value.
//
// Returns:
// - genesis.FullTick: A struct containing the tick index, and tick information.
// - error: An error if the key or value is not valid or if the parsing fails.
func ParseFullTickFromBytes(key, value []byte) (tick genesis.FullTick, err error) {
	if len(key) == 0 {
		return genesis.FullTick{}, types.ErrKeyNotFound
	}
	if len(value) == 0 {
		return genesis.FullTick{}, types.ValueNotFoundForKeyError{Key: key}
	}

	if len(key) != types.KeyTickLengthBytes {
		return genesis.FullTick{}, types.InvalidTickKeyByteLengthError{Length: len(key)}
	}

	prefix := key[0:len(types.TickPrefix)]
	if !bytes.Equal(types.TickPrefix, prefix) {
		return genesis.FullTick{}, types.InvalidPrefixError{Actual: string(prefix), Expected: string(types.TickPrefix)}
	}

	key = key[len(types.TickPrefix):]

	// We only care about the last componenent, which is the tick index
	key = key[uint64Bytes:]

	tickIndex, err := types.TickIndexFromBytes(key)
	if err != nil {
		return genesis.FullTick{}, err
	}

	tickInfo, err := ParseTickFromBzAndRemoveUnInitializedUptimeTrackers(value)
	if err != nil {
		return genesis.FullTick{}, types.ValueParseError{Wrapped: err}
	}

	return genesis.FullTick{
		TickIndex: tickIndex,
		Info:      tickInfo,
	}, nil
}
