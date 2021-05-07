package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/pool-incentives/types"
)

func (k Keeper) HandleUpdatePoolIncentivesProposal(ctx sdk.Context, p *types.UpdatePoolIncentivesProposal) error {
	return k.UpdateDistrRecords(ctx, p.Records...)
}
