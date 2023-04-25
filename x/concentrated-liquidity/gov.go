package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

// HandleTickSpacingDecreaseProposal handles a tick spacing decrease proposal to the corresponding keeper method.
func (k Keeper) HandleTickSpacingDecreaseProposal(ctx sdk.Context, p *types.TickSpacingDecreaseProposal) error {
	return k.DecreaseConcentratedPoolTickSpacing(ctx, p.PoolIdToTickSpacingRecords)
}
