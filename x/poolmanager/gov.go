package poolmanager

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

func (k Keeper) HandleDenomPairTakerFeeProposal(ctx sdk.Context, p *types.DenomPairTakerFeeProposal) error {
	for _, denomPair := range p.DenomPairTakerFee {
		k.SetDenomPairTakerFee(ctx, denomPair.TokenInDenom, denomPair.TokenOutDenom, denomPair.TakerFee)
	}
	return nil
}

func NewPoolManagerProposalHandler(k Keeper) govtypesv1.Handler {
	return func(ctx sdk.Context, content govtypesv1.Content) error {
		switch c := content.(type) {
		case *types.DenomPairTakerFeeProposal:
			return k.HandleDenomPairTakerFeeProposal(ctx, c)

		default:
			return fmt.Errorf("unrecognized pool manager proposal content type: %T", c)
		}
	}
}
