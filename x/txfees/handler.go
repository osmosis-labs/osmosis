package txfees

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v7/x/txfees/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

func NewUpdateFeeTokenProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.UpdateFeeTokenProposal:
			return handleUpdateFeeTokenProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized txfees proposal content type: %T", c)
		}
	}
}

func handleUpdateFeeTokenProposal(ctx sdk.Context, k keeper.Keeper, p *types.UpdateFeeTokenProposal) error {
	return k.HandleUpdateFeeTokenProposal(ctx, p)
}
