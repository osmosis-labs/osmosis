package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/pool-yield/types"
)

func (k Keeper) HandleAddPoolIncentivesProposal(ctx sdk.Context, p *types.AddPoolIncentivesProposal) error {
	return k.AddDistrRecords(ctx, p.Records...)
}

func (k Keeper) HandleRemovePoolIncentivesProposal(ctx sdk.Context, p *types.RemovePoolIncentivesProposal) error {
	var indexes []int
	for _, index := range p.Indexes {
		indexes = append(indexes, int(index))
	}
	k.RemoveDistrRecords(ctx, indexes...)

	return nil
}
