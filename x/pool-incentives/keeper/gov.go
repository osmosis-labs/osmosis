package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v25/x/pool-incentives/types"
)

func (k Keeper) HandleReplacePoolIncentivesProposal(ctx sdk.Context, p *types.ReplacePoolIncentivesProposal) error {
	return k.ReplaceDistrRecords(ctx, p.Records...)
}

// TODO: Remove in v27 once comfortable with new gov message
func (k Keeper) HandleUpdatePoolIncentivesProposal(ctx sdk.Context, p *types.UpdatePoolIncentivesProposal) error {
	return k.UpdateDistrRecords(ctx, p.Records...)
}
