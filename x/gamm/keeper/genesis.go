package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/types"
)

<<<<<<< HEAD
// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState, unpacker codectypes.AnyUnpacker) {
	k.SetParams(ctx, genState.Params)
	k.SetNextPoolNumber(ctx, genState.NextPoolNumber)
=======
// InitGenesis initializes the x/gamm module's state from a provided genesis
// state, which includes the current live pools, global pool parameters (e.g. pool creation fee), next pool id etc.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState, unpacker codectypes.AnyUnpacker) {
	k.setParams(ctx, genState.Params)
	k.setNextPoolId(ctx, genState.NextPoolNumber)
>>>>>>> 8cac0906 (gamm keeper delete obsolete files (#2160))

	liquidity := sdk.Coins{}
	for _, any := range genState.Pools {
		var pool types.PoolI
		err := unpacker.UnpackAny(any, &pool)
		if err != nil {
			panic(err)
		}
		err = k.SetPool(ctx, pool)
		if err != nil {
			panic(err)
		}

		poolAssets := pool.GetTotalPoolLiquidity(ctx)
		for _, asset := range poolAssets {
			liquidity = liquidity.Add(asset)
		}
	}

	k.SetTotalLiquidity(ctx, liquidity)
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
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
<<<<<<< HEAD
		NextPoolNumber: k.GetNextPoolNumberAndIncrement(ctx),
=======
		NextPoolNumber: k.GetNextPoolId(ctx),
>>>>>>> 8cac0906 (gamm keeper delete obsolete files (#2160))
		Pools:          poolAnys,
		Params:         k.GetParams(ctx),
	}
}
