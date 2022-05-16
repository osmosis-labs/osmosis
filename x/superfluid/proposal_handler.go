package superfluid

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v8/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v8/x/superfluid/keeper/gov"
	"github.com/osmosis-labs/osmosis/v8/x/superfluid/types"
)

func NewSuperfluidProposalHandler(k keeper.Keeper, ek types.EpochKeeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.SetSuperfluidAssetsProposal:
			return handleSetSuperfluidAssetsProposal(ctx, k, ek, c)
		case *types.RemoveSuperfluidAssetsProposal:
			return handleRemoveSuperfluidAssetsProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized pool incentives proposal content type: %T", c)
		}
	}
}

func handleSetSuperfluidAssetsProposal(ctx sdk.Context, k keeper.Keeper, ek types.EpochKeeper, p *types.SetSuperfluidAssetsProposal) error {
	return gov.HandleSetSuperfluidAssetsProposal(ctx, k, ek, p)
}

func handleRemoveSuperfluidAssetsProposal(ctx sdk.Context, k keeper.Keeper, p *types.RemoveSuperfluidAssetsProposal) error {
	return gov.HandleRemoveSuperfluidAssetsProposal(ctx, k, p)
}
