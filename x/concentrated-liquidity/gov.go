package concentrated_liquidity

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/osmosis/v29/x/concentrated-liquidity/types"
)

// HandleTickSpacingDecreaseProposal handles a tick spacing decrease proposal to the corresponding keeper method.
func (k Keeper) HandleTickSpacingDecreaseProposal(ctx sdk.Context, p *types.TickSpacingDecreaseProposal) error {
	return k.DecreaseConcentratedPoolTickSpacing(ctx, p.PoolIdToTickSpacingRecords)
}

func NewConcentratedLiquidityProposalHandler(k Keeper) govtypesv1.Handler {
	return func(ctx sdk.Context, content govtypesv1.Content) error {
		switch c := content.(type) {
		case *types.TickSpacingDecreaseProposal:
			return k.HandleTickSpacingDecreaseProposal(ctx, c)
		default:
			return fmt.Errorf("unrecognized concentrated liquidity proposal content type: %T", c)
		}
	}
}
