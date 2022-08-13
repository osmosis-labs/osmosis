package pool_incentives

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v11/x/pool-incentives/keeper"
	"github.com/osmosis-labs/osmosis/v11/x/pool-incentives/types"
)

func NewPoolIncentivesProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.UpdatePoolIncentivesProposal:
			return handleUpdatePoolIncentivesProposal(ctx, k, c)
		case *types.ReplacePoolIncentivesProposal:
			return handleReplacePoolIncentivesProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized pool incentives proposal content type: %T", c)
		}
	}
}

func handleReplacePoolIncentivesProposal(ctx sdk.Context, k keeper.Keeper, p *types.ReplacePoolIncentivesProposal) error {
	return k.HandleReplacePoolIncentivesProposal(ctx, p)
}

func handleUpdatePoolIncentivesProposal(ctx sdk.Context, k keeper.Keeper, p *types.UpdatePoolIncentivesProposal) error {
	return k.HandleUpdatePoolIncentivesProposal(ctx, p)
}
