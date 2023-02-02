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
	k.SetPointCountForBlock(ctx, 0)

	// Configure max pool points per block. This roughly correlates to the ms of execution time protorev will
	// take per block
	if err := k.SetMaxPointsPerBlock(ctx, 100); err != nil {
		panic(err)
	}

	// Configure max pool points per tx. This roughly correlates to the ms of execution time protorev will take
	// per tx
	if err := k.SetMaxPointsPerTx(ctx, 18); err != nil {
		panic(err)
	}

	// Configure the pool weights for genesis. This roughly correlates to the ms of execution time
	// by pool type
	poolWeights := types.PoolWeights{
		StableWeight:       5, // it takes around 5 ms to simulate and execute a stable swap
		BalancerWeight:     2, // it takes around 2 ms to simulate and execute a balancer swap
		ConcentratedWeight: 2, // it takes around 2 ms to simulate and execute a concentrated swap
	}
	k.SetPoolWeights(ctx, poolWeights)

	// Configure the initial base denoms used for cyclic route building
	baseDenomPriorities := []string{types.OsmosisDenomination, types.AtomDenomination}
	k.SetBaseDenoms(ctx, baseDenomPriorities)

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
