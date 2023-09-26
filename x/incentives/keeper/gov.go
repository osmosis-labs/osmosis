package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v19/x/incentives/types"
)

func (k Keeper) HandleCreateGaugeGroupsProposal(ctx sdk.Context, p *types.CreateGaugeGroupsProposal) error {
	for _, group := range p.CreateGroups {
		k.CreateGroup(ctx, group.Coins, group.NumEpochsPaidOver, group.FilledEpochs, group.PoolIds)
	}
	return nil
}

func NewIncentivesProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.CreateGaugeGroupsProposal:
			return k.HandleCreateGaugeGroupsProposal(ctx, c)

		default:
			return fmt.Errorf("unrecognized incentives proposal content type: %T", c)
		}
	}
}
