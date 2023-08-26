package poolmanager

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

func (k Keeper) HandleDenomPairTakerFeeProposal(ctx sdk.Context, p *types.DenomPairTakerFeeProposal) error {
	for _, denomPair := range p.DenomPairTakerFee {
		k.SetDenomPairTakerFee(ctx, denomPair.Denom0, denomPair.Denom1, denomPair.TakerFee)
	}
	return nil
}

func NewPoolManagerProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.DenomPairTakerFeeProposal:
			return k.HandleDenomPairTakerFeeProposal(ctx, c)

		default:
			return fmt.Errorf("unrecognized pool manager proposal content type: %T", c)
		}
	}
}
