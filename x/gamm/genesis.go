package gamm

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState, unpacker codectypes.AnyUnpacker) {
	k.SetParams(ctx, genState.Params)
	k.SetNextPoolNumber(ctx, genState.NextPoolNumber)

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

		poolAssets := pool.GetAllPoolAssets()
		for _, asset := range poolAssets {
			liquidity = liquidity.Add(asset.Token)
		}
	}

	k.SetTotalLiquidity(ctx, liquidity)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	pools, err := k.GetPools(ctx)
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
