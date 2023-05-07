package concentrated_liquidity

import (
	"errors"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	types "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/genesis"
)

// InitGenesis initializes the concentrated-liquidity module with the provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState genesis.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.SetNextPositionId(ctx, genState.NextPositionId)
	// Initialize pools
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
			k.SetTickInfo(ctx, poolId, tick.TickIndex, tick.Info)
		}
		seenPoolIds[poolId] = struct{}{}

		// set up fee accumulators
		store := ctx.KVStore(k.storeKey)
		err = accum.MakeAccumulatorWithValueAndShare(store, poolData.FeeAccumulator.Name, poolData.FeeAccumulator.AccumContent.AccumValue, poolData.FeeAccumulator.AccumContent.TotalShares)
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

		// set individual fee accumulator state position
		feeAccumObject, err := k.GetFeeAccumulator(ctx, positionWrapper.Position.PoolId)
		if err != nil {
			panic(err)
		}
		feePositionKey := types.KeyFeePositionAccumulator(positionWrapper.Position.PositionId)

		k.initOrUpdateAccumPosition(ctx, feeAccumObject, positionWrapper.FeeAccumRecord.InitAccumValue, feePositionKey, positionWrapper.FeeAccumRecord.NumShares, positionWrapper.FeeAccumRecord.UnclaimedRewards, positionWrapper.FeeAccumRecord.Options)
	}
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
		accumObject, err := k.GetFeeAccumulator(ctx, poolI.GetId())
		if err != nil {
			panic(err)
		}

		totalShares, err := accumObject.GetTotalShares()
		if err != nil {
			panic(err)
		}

		feeAccumObject := genesis.AccumObject{
			Name: types.KeyFeePoolAccumulator(poolI.GetId()),
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

		poolData = append(poolData, genesis.PoolData{
			Pool:                   &anyCopy,
			Ticks:                  ticks,
			FeeAccumulator:         feeAccumObject,
			IncentivesAccumulators: incentivesAccumObject,
			IncentiveRecords:       incentiveRecordsForPool,
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
			if errors.Is(err, types.PositionIdToLockNotFoundError{PositionId: position.PositionId}) {
				lockId = 0
			} else {
				panic(err)
			}
		}

		// Retrieve fee accumulator state for position
		feePositionKey := types.KeyFeePositionAccumulator(position.PositionId)
		feeAccumObject, err := k.GetFeeAccumulator(ctx, position.PoolId)
		if err != nil {
			panic(err)
		}
		feeAccumPositionRecord, err := feeAccumObject.GetPosition(feePositionKey)
		if err != nil {
			panic(err)
		}

		positionData = append(positionData, genesis.PositionData{
			LockId:         lockId,
			Position:       &position,
			FeeAccumRecord: feeAccumPositionRecord,
		})
	}

	return &genesis.GenesisState{
		Params:         k.GetParams(ctx),
		PoolData:       poolData,
		PositionData:   positionData,
		NextPositionId: k.GetNextPositionId(ctx),
	}
}

// initOrUpdateAccumPosition creates a new position or override an existing position
// at accumulator's current value with a specific number of shares and unclaimed rewards
func (k Keeper) initOrUpdateAccumPosition(ctx sdk.Context, accumumulator accum.AccumulatorObject, accumulatorValue sdk.DecCoins, index string, numShareUnits sdk.Dec, unclaimedRewards sdk.DecCoins, options *accum.Options) {
	position := accum.Record{
		NumShares:        numShareUnits,
		InitAccumValue:   accumulatorValue,
		UnclaimedRewards: unclaimedRewards,
		Options:          options,
	}

	osmoutils.MustSet(ctx.KVStore(k.storeKey), accum.FormatPositionPrefixKey(accumumulator.GetName(), index), &position)
}
