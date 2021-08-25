package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (k Keeper) HandleSetSuperfluidAssetsProposal(ctx sdk.Context, p *types.SetSuperfluidAssetsProposal) error {
	return nil
}

func (k Keeper) HandleEnableSuperfluidAssetsProposal(ctx sdk.Context, p *types.EnableSuperfluidAssetsProposal) error {
	return nil
}

func (k Keeper) HandleDisableSuperfluidAssetsProposal(ctx sdk.Context, p *types.DisableSuperfluidAssetsProposal) error {
	return nil
}
