package superfluid

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// initialize superfluid assets
	for _, asset := range genState.SuperfluidAssets {
		k.SetSuperfluidAsset(ctx, asset)
	}

	// initialize superfluid asset infos
	for _, assetInfo := range genState.SuperfluidAssetInfos {
		k.SetSuperfluidAssetInfo(ctx, assetInfo)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {

	return &types.GenesisState{
		SuperfluidAssets:     k.GetAllSuperfluidAssets(ctx),
		SuperfluidAssetInfos: k.GetAllSuperfluidAssetInfos(ctx),
	}
}
