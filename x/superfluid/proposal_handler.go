package superfluid

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper/gov"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

func NewSuperfluidProposalHandler(k keeper.Keeper, ek types.EpochKeeper, gk types.GammKeeper) govtypesv1.Handler {
	return func(ctx sdk.Context, content govtypesv1.Content) error {
		switch c := content.(type) {
		case *types.SetSuperfluidAssetsProposal:
			return handleSetSuperfluidAssetsProposal(ctx, k, ek, c)
		case *types.RemoveSuperfluidAssetsProposal:
			return handleRemoveSuperfluidAssetsProposal(ctx, k, c)
		case *types.UpdateUnpoolWhiteListProposal:
			return handleUnpoolWhitelistChange(ctx, k, gk, c)

		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized pool incentives proposal content type: %T", c)
		}
	}
}

func handleSetSuperfluidAssetsProposal(ctx sdk.Context, k keeper.Keeper, ek types.EpochKeeper, p *types.SetSuperfluidAssetsProposal) error {
	return gov.HandleSetSuperfluidAssetsProposal(ctx, k, ek, p)
}

func handleRemoveSuperfluidAssetsProposal(ctx sdk.Context, k keeper.Keeper, p *types.RemoveSuperfluidAssetsProposal) error {
	return gov.HandleRemoveSuperfluidAssetsProposal(ctx, k, p)
}

func handleUnpoolWhitelistChange(ctx sdk.Context, k keeper.Keeper, gammKeeper types.GammKeeper, p *types.UpdateUnpoolWhiteListProposal) error {
	return gov.HandleUnpoolWhiteListChange(ctx, k, gammKeeper, p)
}
