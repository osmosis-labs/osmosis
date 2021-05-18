package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	params := k.GetParams(ctx)
	if epochIdentifier == params.DistrEpochIdentifier {
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
	}
}

//____________________________________________________________________________________________________

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
