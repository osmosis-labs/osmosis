package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
)

// InitGenesis initializes the x/gamm module's state from a provided genesis
// state, which includes the current live pools, global pool parameters (e.g. pool creation fee), next pool id etc.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState, unpacker codectypes.AnyUnpacker) {
	k.setParams(ctx, genState.Params)
	k.setNextPoolId(ctx, genState.NextPoolNumber)

	// Sums up the liquidity in all genesis state pools to find the total liquidity across all pools.
	// Also adds each genesis state pool to the x/gamm module's state
	liquidity := sdk.Coins{}
	for _, any := range genState.Pools {
		var pool types.CFMMPoolI
		err := unpacker.UnpackAny(any, &pool)
		if err != nil {
			panic(err)
		}
		err = k.setPool(ctx, pool)
		if err != nil {
			panic(err)
		}

		poolAssets := pool.GetTotalPoolLiquidity(ctx)
		for _, asset := range poolAssets {
			liquidity = liquidity.Add(asset)
		}
	}

	k.setTotalLiquidity(ctx, liquidity)

	if genState.MigrationRecords == nil {
		k.SetMigrationInfo(ctx, types.MigrationRecords{})
	} else {
		k.SetMigrationInfo(ctx, *genState.MigrationRecords)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	migrationInfo := k.GetMigrationInfo(ctx)
	pools, err := k.GetPoolsAndPoke(ctx)
	if err != nil {
		panic(err)
	}
	poolAnys := []*codectypes.Any{}
	for _, poolI := range pools {
		any, err := codectypes.NewAnyWithValue(poolI)
		if err != nil {
			panic(err)
		}
		poolAnys = append(poolAnys, any)
	}
	return &types.GenesisState{
		NextPoolNumber:   k.GetNextPoolId(ctx),
		Pools:            poolAnys,
		Params:           k.GetParams(ctx),
		MigrationRecords: &migrationInfo,
	}
}
