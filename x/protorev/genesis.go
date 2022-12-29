package protorev

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Perform stateless validation on the genesis state
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	// Init module state
	k.SetParams(ctx, genState.Params)
	k.SetProtoRevEnabled(ctx, genState.Params.Enabled)
	k.SetDaysSinceModuleGenesis(ctx, 0)
	k.SetLatestBlockHeight(ctx, uint64(ctx.BlockHeight()))
	k.SetRouteCountForBlock(ctx, 0)

	// configure max routes per block (default 100)
	if err := k.SetMaxRoutesPerBlock(ctx, 100); err != nil {
		panic(err)
	}

	// configure max routes per tx (default 6)
	if err := k.SetMaxRoutesPerTx(ctx, 6); err != nil {
		panic(err)
	}

	// Update the pools on genesis
	if err := k.UpdatePools(ctx); err != nil {
		panic(err)
	}

	// Init all of the searcher routes
	for _, tokenPairArbRoutes := range genState.TokenPairs {
		err := tokenPairArbRoutes.Validate()
		if err != nil {
			panic(err)
		}

		_, err = k.SetTokenPairArbRoutes(ctx, tokenPairArbRoutes.TokenIn, tokenPairArbRoutes.TokenOut, &tokenPairArbRoutes)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	return genesis
}
