package txfees

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/osmosis/v31/x/txfees/keeper"
	"github.com/osmosis-labs/osmosis/v31/x/txfees/types"
)

func NewUpdateFeeTokenProposalHandler(k keeper.Keeper) govtypesv1.Handler {
	return func(ctx sdk.Context, content govtypesv1.Content) error {
		switch c := content.(type) {
		case *types.UpdateFeeTokenProposal:
			return handleUpdateFeeTokenProposal(ctx, k, c)
		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized txfees proposal content type: %T", c)
		}
	}
}

func handleUpdateFeeTokenProposal(ctx sdk.Context, k keeper.Keeper, p *types.UpdateFeeTokenProposal) error {
	return k.HandleUpdateFeeTokenProposal(ctx, p)
}
