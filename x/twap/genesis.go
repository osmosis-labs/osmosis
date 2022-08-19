package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)

	for _, twap := range genState.Twaps {
		k.storeNewRecord(ctx, twap)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	twapRecords, err := k.getAllHistoricalPoolIndexedTWAPs(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params: k.GetParams(ctx),
		Twaps:  twapRecords,
	}
}
