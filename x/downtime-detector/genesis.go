package downtimedetector

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/downtime-detector/types"
)

func (k *Keeper) InitGenesis(ctx sdk.Context, gen *types.GenesisState) {

}

// ExportGenesis returns the twap module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		// Params: k.GetParams(ctx),
	}
}
