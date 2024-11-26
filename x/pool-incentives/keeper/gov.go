package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
)

func (k Keeper) HandleReplacePoolIncentivesProposal(ctx sdk.Context, p *types.ReplacePoolIncentivesProposal) error {
	return k.ReplaceDistrRecords(ctx, p.Records...)
}

func (k Keeper) HandleUpdatePoolIncentivesProposal(ctx sdk.Context, p *types.UpdatePoolIncentivesProposal) error {
	return k.UpdateDistrRecords(ctx, p.Records...)
}
