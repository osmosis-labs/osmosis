package concentrated_liquidity

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

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

		// set up accumulators
		store := ctx.KVStore(k.storeKey)
		err = accum.MakeAccumulatorWithValueAndShare(store, poolData.AccumObject.Name, poolData.AccumObject.Value, poolData.AccumObject.TotalShares)
		if err != nil {
			panic(err)
		}
	}

	// set incentive records
	err := k.setMultipleIncentiveRecords(ctx, genState.IncentiveRecords)
	if err != nil {
		panic(err)
	}

	// set positions for pool
	for _, position := range genState.Positions {
		if _, ok := seenPoolIds[position.PoolId]; !ok {
			panic(fmt.Sprintf("found position with pool id (%d) but there is no pool with such id that exists", position.PoolId))
		}
		k.setPosition(ctx, position.PoolId, sdk.MustAccAddressFromBech32(position.Address), position.LowerTick, position.UpperTick, position.JoinTime, position.FreezeDuration, position.Liquidity)
	}
}

// ExportGenesis returns the concentrated-liquidity module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *genesis.GenesisState {
	pools, err := k.GetPools(ctx)
	if err != nil {
		panic(err)
	}

	poolData := make([]genesis.PoolData, 0, len(pools))
	incentiveRecords := []types.IncentiveRecord{}

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

		genAccumObject := genesis.AccumObject{
			Name:        getFeeAccumulatorName(poolI.GetId()),
			Value:       accumObject.GetValue(),
			TotalShares: totalShares,
		}

		poolData = append(poolData, genesis.PoolData{
			Pool:        &anyCopy,
			Ticks:       ticks,
			AccumObject: genAccumObject,
		})

		poolId := poolI.GetId()
		incentiveRecordsForPool, err := k.GetAllIncentiveRecordsForPool(ctx, poolId)
		if err != nil {
			panic(err)
		}
		incentiveRecords = append(incentiveRecords, incentiveRecordsForPool...)
	}

	positions, err := k.getAllPositions(ctx)
	if err != nil {
		panic(err)
	}

	return &genesis.GenesisState{
		Params:           k.GetParams(ctx),
		PoolData:         poolData,
		Positions:        positions,
		IncentiveRecords: incentiveRecords,
		NextPositionId:   k.GetNextPositionId(ctx),
	}
}
