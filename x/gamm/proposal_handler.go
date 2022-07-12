package gamm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper/gov"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func NewGAMMProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.SetSwapFeeProposal:
			return handleSetSwapFeeProposal(ctx, k, c)
		case *types.SetExitFeeProposal:
			return handleSetExitFeeProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized gamm proposal content type: %T", c)
		}
	}
}

func handleSetSwapFeeProposal(ctx sdk.Context, k keeper.Keeper, p *types.SetSwapFeeProposal) error {
	return gov.HandleSetSwapFeeProposal(ctx, k, p)
}

func handleSetExitFeeProposal(ctx sdk.Context, k keeper.Keeper, p *types.SetExitFeeProposal) error {
	return gov.HandleSetExitFeeProposal(ctx, k, p)
}
