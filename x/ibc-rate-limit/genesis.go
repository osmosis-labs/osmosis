package ibc_rate_limit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/types"
)

// InitGenesis initializes the x/ibc-rate-limit module's state from a provided genesis
// state, which includes the parameter for the contract address.
func (i *ICS4Wrapper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	i.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the x/ibc-rate-limit module's exported genesis.
func (i *ICS4Wrapper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: i.GetParams(ctx),
	}
}
