package gamm

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetNextPoolNumber(ctx, genState.NextPoolNumber)

	liquidity := sdk.Coins{}
	for _, pool := range genState.Pools {
		pool := pool.GetCachedValue().(types.PoolI)
		err := k.SetPool(ctx, pool)
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
	poolsAny := []*codectypes.Any{}
	for _, pool := range pools {
		any, err := codectypes.NewAnyWithValue(pool)
		if err != nil {
			panic(err)
		}
		poolsAny = append(poolsAny, any)
	}
	return &types.GenesisState{
		NextPoolNumber: k.GetNextPoolNumber(ctx),
		Pools:          poolsAny,
	}
}
