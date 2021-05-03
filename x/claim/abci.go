package claim

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/claim/keeper"
)

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {

	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	// Only check whether to end airdrop if it hasn't already ended
	if !k.GetAirdropCompleted(ctx) {
		// End Airdrop
		goneTime := ctx.BlockTime().Sub(params.AirdropStart)
		if goneTime > params.DurationUntilDecay+params.DurationOfDecay {
			// airdrop time passed
			k.EndAirdrop(ctx)
		}
	}
}
