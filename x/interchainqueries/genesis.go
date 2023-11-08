package interchainqueries

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/x/interchainqueries/keeper"
	"github.com/osmosis-labs/osmosis/v20/x/interchainqueries/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	sort.SliceStable(genState.RegisteredQueries, func(i, j int) bool {
		return genState.RegisteredQueries[i].Id < genState.RegisteredQueries[j].Id
	})

	// Set all registered queries
	for _, elem := range genState.RegisteredQueries {
		k.SetLastRegisteredQueryKey(ctx, elem.Id)
		if err := k.SaveQuery(ctx, elem); err != nil {
			panic(err)
		}
	}

	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.RegisteredQueries = k.GetAllRegisteredQueries(ctx)

	return genesis
}
