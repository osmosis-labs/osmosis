package superfluid

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func NewSuperfluidProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.SetSuperfluidAssetsProposal:
			return handleSetSuperfluidAssetsProposal(ctx, k, c)
		case *types.EnableSuperfluidAssetsProposal:
			return handleEnableSuperfluidAssetsProposal(ctx, k, c)
		case *types.DisableSuperfluidAssetsProposal:
			return handleDisableSuperfluidAssetsProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized pool incentives proposal content type: %T", c)
		}
	}
}

func handleSetSuperfluidAssetsProposal(ctx sdk.Context, k keeper.Keeper, p *types.SetSuperfluidAssetsProposal) error {
	return k.HandleSetSuperfluidAssetsProposal(ctx, p)
}

func handleEnableSuperfluidAssetsProposal(ctx sdk.Context, k keeper.Keeper, p *types.EnableSuperfluidAssetsProposal) error {
	return k.HandleEnableSuperfluidAssetsProposal(ctx, p)
}

func handleDisableSuperfluidAssetsProposal(ctx sdk.Context, k keeper.Keeper, p *types.DisableSuperfluidAssetsProposal) error {
	return k.HandleDisableSuperfluidAssetsProposal(ctx, p)
}
