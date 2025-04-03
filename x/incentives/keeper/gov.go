package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
)

func (k Keeper) HandleCreateGaugeProposal(ctx sdk.Context, p *types.CreateGroupsProposal) error {
	for _, group := range p.CreateGroups {
		incentivesModuleAddress := k.ak.GetModuleAddress(types.ModuleName)
		// N.B: We force internal gauge creation here only because we don't have a straightforward
		// way to escrow the funds from the prop creator to be used at time of prop execution (or returned if the prop fails).
		// Once we have a way to do this, we can change the CreateGroups proto to allow for coins and numEpochsPaidOver and
		// then modify it here as well.
		// Note: do not replace with CreateGroupAsIncentivesModuleAcc as that implementation does not attempt to sync weights
		// We still want to sync the weights here to ensure that the pools are valid and have the associated volume at group creation time.
		_, err := k.CreateGroup(ctx, sdk.Coins{}, types.PerpetualNumEpochsPaidOver, incentivesModuleAddress, group.PoolIds)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewIncentivesProposalHandler(k Keeper) govtypesv1.Handler {
	return func(ctx sdk.Context, content govtypesv1.Content) error {
		switch c := content.(type) {
		case *types.CreateGroupsProposal:
			return k.HandleCreateGaugeProposal(ctx, c)

		default:
			return fmt.Errorf("unrecognized incentives proposal content type: %T", c)
		}
	}
}
