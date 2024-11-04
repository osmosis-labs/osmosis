package pool_incentives

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
)

// NewPoolIncentivesProposalHandler is a handler for governance proposals on new pool incentives.
func NewPoolIncentivesProposalHandler(k keeper.Keeper) govtypesv1.Handler {
	return func(ctx sdk.Context, content govtypesv1.Content) error {
		switch c := content.(type) {
		case *types.UpdatePoolIncentivesProposal:
			return handleUpdatePoolIncentivesProposal(ctx, k, c)
		case *types.ReplacePoolIncentivesProposal:
			return handleReplacePoolIncentivesProposal(ctx, k, c)

		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized pool incentives proposal content type: %T", c)
		}
	}
}

// handleReplacePoolIncentivesProposal is a handler for replacing pool incentives governance proposals
func handleReplacePoolIncentivesProposal(ctx sdk.Context, k keeper.Keeper, p *types.ReplacePoolIncentivesProposal) error {
	return k.HandleReplacePoolIncentivesProposal(ctx, p)
}

// handleUpdatePoolIncentivesProposal is a handler for updating pool incentives governance proposals
func handleUpdatePoolIncentivesProposal(ctx sdk.Context, k keeper.Keeper, p *types.UpdatePoolIncentivesProposal) error {
	return k.HandleUpdatePoolIncentivesProposal(ctx, p)
}
