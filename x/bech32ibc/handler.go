package bech32ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/x/bech32ibc/keeper"
	"github.com/osmosis-labs/osmosis/x/bech32ibc/types"
)

// NewHandler returns claim module messages
func NewHandler(k keeper.Keeper) sdk.Handler {

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {

		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

func NewPoolIncentivesProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.UpdateHrpIbcChannelProposal:
			return handleUpdateHrpIbcChannelProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized bech32 ibc proposal content type: %T", c)
		}
	}
}

func handleUpdateHrpIbcChannelProposal(ctx sdk.Context, k keeper.Keeper, p *types.UpdateHrpIbcChannelProposal) error {
	return k.HandleUpdateHrpIbcChannelProposal(ctx, p)
}
