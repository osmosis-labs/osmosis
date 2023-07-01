package cosmwasmpool

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool/types"
)

// InitGenesis initializes the store state from a genesis state.
func (k *Keeper) InitGenesis(ctx sdk.Context, gen *types.GenesisState, unpacker codectypes.AnyUnpacker) {
	k.SetParams(ctx, gen.Params)

	// Sums up the liquidity in all genesis state pools to find the total liquidity across all pools.
	// Also adds each genesis state pool to the x/gamm module's state
	liquidity := sdk.Coins{}
	for _, any := range gen.Pools {
		var pool types.CosmWasmExtension
		err := unpacker.UnpackAny(any, &pool)
		if err != nil {
			panic(err)
		}
		k.SetPool(ctx, pool)

		poolAssets := pool.GetTotalPoolLiquidity(ctx)
		for _, asset := range poolAssets {
			liquidity = liquidity.Add(asset)
		}
	}

	k.setTotalLiquidity(ctx, liquidity)
}

// ExportGenesis returns the cosmwasm pool's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)

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
		Params: params,
		Pools:  poolAnys,
	}
}
