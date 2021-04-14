package incentives

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/c-osmosis/osmosis/x/incentives/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker is called on every block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
}

// Called every block to distribute coins
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	currentEpoch, epochBeginBlock := k.GetCurrentEpochInfo(ctx)
	params := k.GetParams(ctx)

	if ctx.BlockHeight() < epochBeginBlock+params.BlocksPerEpoch { // not time for epoch
		return []abci.ValidatorUpdate{}
	}

	// update epoch info
	k.SetCurrentEpochInfo(ctx, currentEpoch+1, ctx.BlockHeight())

	// begin distribution if it's start time
	pots := k.GetUpcomingPots(ctx)
	for _, pot := range pots {
		if pot.StartTime.Before(ctx.BlockTime()) {
			k.BeginDistribution(ctx, pot)
		}
	}

	// distribute due to epoch event
	pots = k.GetActivePots(ctx)
	for _, pot := range pots {
		k.Distribute(ctx, pot)
		// filled epoch is increased in this step and we compare with +1
		if !pot.IsPerpetual && pot.NumEpochsPaidOver <= pot.FilledEpochs+1 {
			k.FinishDistribution(ctx, pot)
		}
	}

	return []abci.ValidatorUpdate{}
}
