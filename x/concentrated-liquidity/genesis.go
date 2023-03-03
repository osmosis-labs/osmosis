package concentrated_liquidity

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

// InitGenesis initializes the concentrated-liquidity module with the provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	// Initialize pools
	var unpacker codectypes.AnyUnpacker = k.cdc
	for _, poolAny := range genState.Pools {
		var pool types.ConcentratedPoolExtension
		err := unpacker.UnpackAny(poolAny, &pool)
		if err != nil {
			panic(err)
		}
		err = k.setPool(ctx, pool)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the concentrated-liquidity module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	pools, err := k.GetAllPools(ctx)
	if err != nil {
		panic(err)
	}
	poolAnys := []*codectypes.Any{}
	for _, poolI := range pools {
		any, err := codectypes.NewAnyWithValue(poolI)
		if err != nil {
			panic(err)
		}
		anyCopy := *any
		poolAnys = append(poolAnys, &anyCopy)
	}
	return &types.GenesisState{
		Params: k.GetParams(ctx),
		Pools:  poolAnys,
	}
}
