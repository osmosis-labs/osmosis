package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/stable-staking-incentives/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	k.SetParams(ctx, data.Params)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func (k Keeper) ExportGenesis(ctx sdk.Context) (data *types.GenesisState) {
	params := k.GetParams(ctx)
	return types.NewGenesisState(params)
}
