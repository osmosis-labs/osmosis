package ibc_rate_limit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/ibc-rate-limit/types"
)

// InitGenesis initializes the x/ibc-rate-limit module's state from a provided genesis
// state, which includes the current live pools, global pool parameters (e.g. pool creation fee), next pool id etc.
// TODO: test
func (i *ICS4Wrapper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	i.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the x/ibc-rate-limit module's exported genesis.
// TODO: test
func (i *ICS4Wrapper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: i.GetParams(ctx),
	}
}
