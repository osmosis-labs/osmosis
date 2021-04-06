package claim

import (
	"github.com/c-osmosis/osmosis/x/claim/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {

	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	goneTime := ctx.BlockTime().Sub(params.AirdropStart)
	if goneTime > params.DurationUntilDecay+params.DurationOfDecay {
		// airdrop time passed
		k.FundRemainingsToCommunity(ctx)
		k.ClearClaimables(ctx)
	}
}
