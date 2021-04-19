package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/pool-incentives/types"
)

func (k Keeper) HandleAddPoolIncentivesProposal(ctx sdk.Context, p *types.AddPoolIncentivesProposal) error {
	return k.AddDistrRecords(ctx, p.Records...)
}

func (k Keeper) HandleEditPoolIncentivesProposal(ctx sdk.Context, p *types.EditPoolIncentivesProposal) error {
	return k.EditDistrRecords(ctx, p.Records...)
}

func (k Keeper) HandleRemovePoolIncentivesProposal(ctx sdk.Context, p *types.RemovePoolIncentivesProposal) error {
	return k.RemoveDistrRecords(ctx, p.Indexes...)
}
