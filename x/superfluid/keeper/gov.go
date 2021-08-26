package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (k Keeper) HandleSetSuperfluidAssetsProposal(ctx sdk.Context, p *types.SetSuperfluidAssetsProposal) error {
	return nil
}

func (k Keeper) HandleAddSuperfluidAssetsProposal(ctx sdk.Context, p *types.AddSuperfluidAssetsProposal) error {
	return nil
}

func (k Keeper) HandleRemoveSuperfluidAssetsProposal(ctx sdk.Context, p *types.RemoveSuperfluidAssetsProposal) error {
	return nil
}
