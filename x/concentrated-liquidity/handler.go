package concentrated_liquidity

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

func NewConcentratedLiquidityProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.TickSpacingDecreaseProposal:
			return handleTickSpacingDecreaseProposal(ctx, k, c)
		case *types.CreateConcentratedLiquidityPoolProposal:
			return handleCreateConcentratedLiquidityPoolProposal(ctx, k, c)

		default:
			return fmt.Errorf("unrecognized concentrated liquidity proposal content type: %T", c)
		}
	}
}

func handleTickSpacingDecreaseProposal(ctx sdk.Context, k Keeper, p *types.TickSpacingDecreaseProposal) error {
	return k.HandleTickSpacingDecreaseProposal(ctx, p)
}
