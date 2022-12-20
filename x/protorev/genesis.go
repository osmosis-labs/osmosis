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

	// Init module parameters
	k.SetParams(ctx, genState.Params)

	// Init module state
	k.SetProtoRevEnabled(ctx, genState.Params.Enabled)
	k.SetDaysSinceGenesis(ctx, 0)

	// Default we only allow 3 pools to be arbitraged against per tx
	k.SetMaxPools(ctx, 3)

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
