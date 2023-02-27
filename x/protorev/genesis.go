package protorev

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Perform stateless validation on the genesis state
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	// -------------- Init the module state -------------- //
	// Set the module parameters in state
	k.SetParams(ctx, genState.Params)

	// ------------- Route building set up -------------- //
	// Set all of the token pair arb routes in state
	for _, tokenPairArbRoutes := range genState.TokenPairArbRoutes {
		k.SetTokenPairArbRoutes(ctx, tokenPairArbRoutes.TokenIn, tokenPairArbRoutes.TokenOut, tokenPairArbRoutes)
	}

	// Configure the initial base denoms used for cyclic route building. The order of the list of base
	// denoms is the order in which routes will be prioritized i.e. routes will be built and simulated in a
	// first come first serve basis that is based on the order of the base denoms.
	if err := k.SetBaseDenoms(ctx, genState.BaseDenoms); err != nil {
		panic(err)
	}

	// Update the pools on genesis
	if err := k.UpdatePools(ctx); err != nil {
		panic(err)
	}

	// --------------- Developer set up ----------------- //
	// Set the developer fees that have been collected
	for _, fee := range genState.DeveloperFees {
		k.SetDeveloperFees(ctx, fee)
	}

	// Set the number of days since the module genesis
	k.SetDaysSinceModuleGenesis(ctx, genState.DaysSinceModuleGenesis)

	// -------------- Route compute set up -------------- //
	// The current block height should only be updated if the chain has just launched
	k.SetLatestBlockHeight(ctx, genState.LatestBlockHeight)

	// Configure max pool points per tx. This roughly correlates to the ms of execution time protorev will take
	// per tx
	if err := k.SetMaxPointsPerTx(ctx, genState.MaxPoolPointsPerTx); err != nil {
		panic(err)
	}

	// Configure max pool points per block. This roughly correlates to the ms of execution time protorev will
	// take per block
	if err := k.SetMaxPointsPerBlock(ctx, genState.MaxPoolPointsPerBlock); err != nil {
		panic(err)
	}

	// Set the number of pool points that have been consumed in the current block
	k.SetPointCountForBlock(ctx, genState.PointCountForBlock)

	// Configure the pool weights for genesis. This roughly correlates to the ms of execution time
	// by pool type
	k.SetPoolWeights(ctx, genState.PoolWeights)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	return genesis
}
