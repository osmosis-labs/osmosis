package concentrated_liquidity

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/genesis"
)

type PositionWrapper struct {
	Position model.Position
	LockId   uint64
}

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
	for _, position := range genState.Positions {
		if _, ok := seenPoolIds[position.PoolId]; !ok {
			panic(fmt.Sprintf("found position with pool id (%d) but there is no pool with such id that exists", position.PoolId))
		}

		pos := PositionWrapper{
			Position: position,
			LockId:   genState.PositionLockId,
		}

		// We hardcode the underlying lock id to 0, because genesisState should already hold the positionId to lockId connections
		err := k.SetPosition(ctx, pos.Position.PoolId, sdk.MustAccAddressFromBech32(pos.Position.Address), pos.Position.LowerTick, pos.Position.UpperTick, pos.Position.JoinTime, pos.Position.Liquidity, pos.Position.PositionId, pos.LockId)
		if err != nil {
			panic(err)
		}
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
		accumObject, err := k.getFeeAccumulator(ctx, poolI.GetId())
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

		incentivesAccum, err := k.getUptimeAccumulators(ctx, poolId)
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

	return &genesis.GenesisState{
		Params:         k.GetParams(ctx),
		PoolData:       poolData,
		Positions:      positions,
		NextPositionId: k.GetNextPositionId(ctx),
	}
}
