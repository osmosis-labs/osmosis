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

	// Configure the initial base denoms used for cyclic route building. The order of the list of base
	// denoms is the order in which routes will be prioritized i.e. routes will be built and simulated in a
	// first come first serve basis that is based on the order of the base denoms.
	baseDenoms := []*types.BaseDenom{
		{
			Denom:    types.OsmosisDenomination,
			StepSize: sdk.NewInt(1_000_000),
		},
	}
	if err := k.SetBaseDenoms(ctx, baseDenoms); err != nil {
		panic(err)
	}

	// Currently configured to be the Skip dev team's address
	// See https://github.com/osmosis-labs/osmosis/issues/4349 for more details
	// Note that governance has full ability to change this live on-chain, and this admin can at most prevent protorev from working.
	// All the settings manager's controls have limits, so it can't lead to a chain halt, excess processing time or prevention of swaps.
	adminAccount, err := sdk.AccAddressFromBech32("osmo17nv67dvc7f8yr00rhgxd688gcn9t9wvhn783z4")
	if err != nil {
		panic(err)
	}
	k.SetAdminAccount(ctx, adminAccount)

	// Update the pools on genesis
	if err := k.UpdatePools(ctx); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	return genesis
}
