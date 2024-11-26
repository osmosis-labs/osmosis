package protorev

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v27/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

func NewProtoRevProposalHandler(k keeper.Keeper) govtypesv1.Handler {
	return func(ctx sdk.Context, content govtypesv1.Content) error {
		switch c := content.(type) {
		case *types.SetProtoRevAdminAccountProposal:
			return HandleSetProtoRevAdminAccount(ctx, k, c)
		case *types.SetProtoRevEnabledProposal:
			return HandleEnabledProposal(ctx, k, c)
		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}

// handleSetProtoRevAdminAccount handles a proposal to set the admin account. The admin account has the ability to
// update the hot routes, the developer account, and the number of pools that can be iterated over in a single transaction + block.
func HandleSetProtoRevAdminAccount(ctx sdk.Context, k keeper.Keeper, p *types.SetProtoRevAdminAccountProposal) error {
	// Validate the account address
	account, err := sdk.AccAddressFromBech32(p.Account)
	if err != nil {
		return err
	}

	k.SetAdminAccount(ctx, account)
	return nil
}

// handleEnabledProposal handles a proposal to enable/disable the protorev module.
func HandleEnabledProposal(ctx sdk.Context, k keeper.Keeper, p *types.SetProtoRevEnabledProposal) error {
	k.SetProtoRevEnabled(ctx, p.Enabled)
	return nil
}
