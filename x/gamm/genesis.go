package gamm

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/types"
)

<<<<<<< HEAD:x/gamm/genesis.go
// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState, unpacker codectypes.AnyUnpacker) {
	k.SetParams(ctx, genState.Params)
	k.SetNextPoolNumber(ctx, genState.NextPoolNumber)
=======
// InitGenesis initializes the x/gamm module's state from a provided genesis
// state, which includes the current live pools, global pool parameters (e.g. pool creation fee), next pool number etc.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState, unpacker codectypes.AnyUnpacker) {
	k.setParams(ctx, genState.Params)
	k.setNextPoolNumber(ctx, genState.NextPoolNumber)
>>>>>>> 7fb5f824 (x/gamm: Make all internal set functions private (#2013)):x/gamm/keeper/genesis.go

	liquidity := sdk.Coins{}
	for _, any := range genState.Pools {
		var pool types.PoolI
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
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
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
		NextPoolNumber: k.GetNextPoolNumberAndIncrement(ctx),
		Pools:          poolAnys,
		Params:         k.GetParams(ctx),
	}
}
