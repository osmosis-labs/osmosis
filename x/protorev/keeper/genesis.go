package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// Perform stateless validation on the genesis state.
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	// -------------- Init the module state -------------- //
	// Set the module parameters in state.
	k.SetParams(ctx, genState.Params)

	// ------------- Route building set up -------------- //
	// Set all of the token pair arb routes in state.
	for _, tokenPairArbRoutes := range genState.TokenPairArbRoutes {
		if err := k.SetTokenPairArbRoutes(ctx, tokenPairArbRoutes.TokenIn, tokenPairArbRoutes.TokenOut, tokenPairArbRoutes); err != nil {
			panic(err)
		}
	}

	// Configure the initial base denoms used for cyclic route building. The order of the list of base
	// denoms is the order in which routes will be prioritized i.e. routes will be built and simulated in a
	// first come first serve basis that is based on the order of the base denoms.
	if err := k.SetBaseDenoms(ctx, genState.BaseDenoms); err != nil {
		panic(err)
	}

	// Update the pools on genesis.
	if err := k.UpdatePools(ctx); err != nil {
		panic(err)
	}

	// --------------- Developer set up ----------------- //
	// Set the developer address if it exists.
	if genState.DeveloperAddress != "" {
		account, err := sdk.AccAddressFromBech32(genState.DeveloperAddress)
		if err != nil {
			panic(err)
		}

		k.SetDeveloperAccount(ctx, account)
	}

	// Set the developer fees that have been collected.
	for _, fee := range genState.DeveloperFees {
		if err := k.SetDeveloperFees(ctx, fee); err != nil {
			panic(err)
		}
	}

	// Set the number of days since the module genesis.
	k.SetDaysSinceModuleGenesis(ctx, genState.DaysSinceModuleGenesis)

	// -------------- Route compute set up -------------- //
	// Set the latest block height ProtoRev has encountered.
	k.SetLatestBlockHeight(ctx, genState.LatestBlockHeight)

	// Configure max pool points per tx. This roughly correlates to the ms of execution time protorev will take
	// per tx.
	if err := k.SetMaxPointsPerTx(ctx, genState.MaxPoolPointsPerTx); err != nil {
		panic(err)
	}

	// Configure max pool points per block. This roughly correlates to the ms of execution time protorev will
	// take per block.
	if err := k.SetMaxPointsPerBlock(ctx, genState.MaxPoolPointsPerBlock); err != nil {
		panic(err)
	}

	// Set the number of pool points that have been consumed in the current block.
	k.SetPointCountForBlock(ctx, genState.PointCountForBlock)

	// Configure the pool info for genesis.
	k.SetInfoByPoolType(ctx, genState.InfoByPoolType)

	// Set the profits that have been collected by Protorev.
	for _, coin := range genState.Profits {
		if err := k.UpdateProfitsByDenom(ctx, coin.Denom, coin.Amount); err != nil {
			panic(err)
		}
	}

	// Since we now track all aspects of protocol revenue, we need to take a snapshot of cyclic arb profits from this module at a certain block height.
	// This allows us to display how much protocol revenue has been generated since block "X" instead of just since the module was initialized.
	if len(genState.CyclicArbTracker.CyclicArb) > 0 {
		k.SetCyclicArbProfitTrackerValue(ctx, genState.CyclicArbTracker.CyclicArb)
	} else {
		k.SetCyclicArbProfitTrackerValue(ctx, genState.Profits)
	}

	if genState.CyclicArbTracker.HeightAccountingStartsFrom != 0 {
		k.SetCyclicArbProfitTrackerStartHeight(ctx, genState.CyclicArbTracker.HeightAccountingStartsFrom)
	} else {
		k.SetCyclicArbProfitTrackerStartHeight(ctx, ctx.BlockHeight())
	}
}

// ExportGenesis returns the module's exported genesis. ExportGenesis intentionally ignores a few of the errors thrown
// by the keeper methods. This is because the keeper methods are only throwing errors if there is an issue unmarshalling
// or if the value had not been set yet (i.e. developer account address). In that case, we just use the default
// values defined in genesis.go in types.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()
	// Export the module parameters.
	genesis.Params = k.GetParams(ctx)

	// Export the pool weights.
	genesis.InfoByPoolType = k.GetInfoByPoolType(ctx)

	// Export the token pair arb routes (hot routes).
	routes, err := k.GetAllTokenPairArbRoutes(ctx)
	if err != nil {
		panic(err)
	}
	genesis.TokenPairArbRoutes = routes

	// Export the base denoms used for cyclic route building.
	baseDenoms, err := k.GetAllBaseDenoms(ctx)
	if err != nil {
		panic(err)
	}
	genesis.BaseDenoms = baseDenoms

	// Export the developer fees that have been collected.
	fees, err := k.GetAllDeveloperFees(ctx)
	if err != nil {
		panic(err)
	}
	genesis.DeveloperFees = fees

	// Export the number of days since module genesis (ignore the case where it has not been set yet).
	if daysSinceGenesis, err := k.GetDaysSinceModuleGenesis(ctx); err == nil {
		genesis.DaysSinceModuleGenesis = daysSinceGenesis
	}

	// Export the developer address (ignore the error in case the developer account was not set yet).
	if developerAddress, err := k.GetDeveloperAccount(ctx); err == nil {
		genesis.DeveloperAddress = developerAddress.String()
	}

	// Export the latest block height (ignore the error in case the latest block height was not set yet).
	if latestBlockHeight, err := k.GetLatestBlockHeight(ctx); err == nil {
		genesis.LatestBlockHeight = latestBlockHeight
	}

	// Export the max pool points per tx (ignore the error in case the max pool points per tx was not set yet).
	if maxPoolPointsPerTx, err := k.GetMaxPointsPerTx(ctx); err == nil {
		genesis.MaxPoolPointsPerTx = maxPoolPointsPerTx
	}

	// Export the max pool points per block (ignore the error in case the max pool points per block was not set yet).
	if maxPoolPointsPerBlock, err := k.GetMaxPointsPerBlock(ctx); err == nil {
		genesis.MaxPoolPointsPerBlock = maxPoolPointsPerBlock
	}

	// Export the number of pool points that have been consumed in the current block (ignore the error in case the
	// point count for block was not set yet).
	if pointCount, err := k.GetPointCountForBlock(ctx); err == nil {
		genesis.PointCountForBlock = pointCount
	}

	// Export the profits that have been collected by Protorev.
	genesis.Profits = k.GetAllProfits(ctx)

	// Export the profits that have been collected by Protorev since a certain block height.
	cyclicArbTracker := types.CyclicArbTracker{
		CyclicArb:                  k.GetCyclicArbProfitTrackerValue(ctx),
		HeightAccountingStartsFrom: k.GetCyclicArbProfitTrackerStartHeight(ctx),
	}
	genesis.CyclicArbTracker = &cyclicArbTracker

	return genesis
}
