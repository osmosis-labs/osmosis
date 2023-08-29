package cosmwasmpool

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v19/x/cosmwasmpool/types"
)

// InitGenesis initializes the store state from a genesis state.
func (k *Keeper) InitGenesis(ctx sdk.Context, gen *types.GenesisState, unpacker codectypes.AnyUnpacker) {
	k.SetParams(ctx, gen.Params)

	// Add each genesis state pool to the x/cosmwasmpool module's state
	for _, any := range gen.Pools {
		var pool types.CosmWasmExtension
		err := unpacker.UnpackAny(any, &pool)
		if err != nil {
			panic(err)
		}
		k.SetPool(ctx, pool)
	}
}

// ExportGenesis returns the cosmwasm pool's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)

	// TODO: We remove this as there is an issue with resolving the
	// type url for the cosmwasm pools.
	// panic: unable to resolve type URL /
	//pools, err := k.GetPools(ctx)
	//if err != nil {
	//	panic(err)
	//}
	//poolAnys := []*codectypes.Any{}
	//for _, poolI := range pools {
	//	cosmwasmPool, ok := poolI.(types.CosmWasmExtension)
	//	if !ok {
	//		panic("invalid pool type")
	//	}
	//	any, err := codectypes.NewAnyWithValue(cosmwasmPool)
	//	if err != nil {
	//		panic(err)
	//	}
	//	poolAnys = append(poolAnys, any)
	//}

	return &types.GenesisState{
		Params: params,
		// TODO: This is likely because amino is being used directly
		// (instead of codec.LegacyAmino which is preferred) or
		// UnpackInterfacesMessage is not defined for some type which
		// contains a protobuf Any either directly or via one of its members.
		// To see a stacktrace of where the error is coming from,
		// set the var Debug = true in codec/types/compat.go
		// Pools:  poolAnys,
	}
}
