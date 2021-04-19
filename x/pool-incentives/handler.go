package pool_incentives

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/c-osmosis/osmosis/x/pool-incentives/keeper"
	"github.com/c-osmosis/osmosis/x/pool-incentives/types"
)

func NewPoolIncentivesProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddPoolIncentivesProposal:
			return handleAddPoolIncentivesProposal(ctx, k, c)

		case *types.EditPoolIncentivesProposal:
			return handleEditPoolIncentivesProposal(ctx, k, c)

		case *types.RemovePoolIncentivesProposal:
			return handleRemovePoolIncentivesProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized pool incentives proposal content type: %T", c)
		}
	}
}

func handleAddPoolIncentivesProposal(ctx sdk.Context, k keeper.Keeper, p *types.AddPoolIncentivesProposal) error {
	return k.HandleAddPoolIncentivesProposal(ctx, p)
}

func handleEditPoolIncentivesProposal(ctx sdk.Context, k keeper.Keeper, p *types.EditPoolIncentivesProposal) error {
	return k.HandleEditPoolIncentivesProposal(ctx, p)
}

func handleRemovePoolIncentivesProposal(ctx sdk.Context, k keeper.Keeper, p *types.RemovePoolIncentivesProposal) error {
	return k.HandleRemovePoolIncentivesProposal(ctx, p)
}
