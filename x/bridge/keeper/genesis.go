package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

// InitGenesis initializes the bridge module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.CreateModuleAccount(ctx)

	// TODO: handle signers creation

	bridgeModuleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	for _, asset := range genState.Params.Assets {
		_, err := k.tokenFactoryKeeper.CreateDenom(ctx, bridgeModuleAddr.String(), asset.Asset.Denom)
		if err != nil {
			panic(fmt.Sprintf("can't create a new denom %s: %s", asset.Asset.Denom, err))
		}
	}

	k.setParams(ctx, genState.Params)
}

// ExportGenesis returns the bridge module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}
