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

	// Configure max routes per block. This roughly correlates to the ms of execution time protorev will
	// take per block
	if err := k.SetMaxRoutesPerBlock(ctx, 100); err != nil {
		panic(err)
	}

	// Configure max routes per tx. This roughly correlates to the ms of execution time protorev will take
	// per tx
	if err := k.SetMaxRoutesPerTx(ctx, 6); err != nil {
		panic(err)
	}

	// Configure the route weights for genesis. This roughly correlates to the ms of execution time
	// by route type
	routeWeights := types.RouteWeights{
		StableWeight:   5, // it takes around 5 ms to execute a stable swap route
		BalancerWeight: 2, // it takes around 2 ms to execute a balancer swap route
	}
	if err := k.SetRouteWeights(ctx, routeWeights); err != nil {
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
