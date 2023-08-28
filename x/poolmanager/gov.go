package poolmanager

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

func (k Keeper) HandleDenomPairTakerFeeProposal(ctx sdk.Context, p *types.DenomPairTakerFeeProposal) error {
	return errors.New("TODO: unimplemented")
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
