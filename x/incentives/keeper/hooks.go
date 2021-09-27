package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	params := k.GetParams(ctx)
	if epochIdentifier == params.DistrEpochIdentifier {
		// begin distribution if it's start time
		gauges := k.GetUpcomingGauges(ctx)
		for _, gauge := range gauges {
			if !ctx.BlockTime().Before(gauge.StartTime) {
				if err := k.BeginDistribution(ctx, gauge); err != nil {
					panic(err)
				}
			}
		}

		// distribute due to epoch event
		ctx.EventManager().IncreaseCapacity(2e6)
		gauges = k.GetActiveGauges(ctx)
		// _, err := k.Distribute(ctx, gauges)
		// if err != nil {
		// 	panic(err)
		// }
		for _, gauge := range gauges {
			err := k.F1Distribute(ctx, gauge)
			if err != nil {
				panic(err)
			}
			// filled epoch is increased in this step and we compare with +1
			if !gauge.IsPerpetual && gauge.NumEpochsPaidOver <= gauge.FilledEpochs+1 {
				if err := k.FinishDistribution(ctx, gauge); err != nil {
					panic(err)
				}
			}
		}
	}
}

// ___________________________________________________________________________________________________

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

//////////////////////////// STH START //////////////////////////////////

var _ lockuptypes.LockupHooks = Hooks{}

func (h Hooks) OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	lockableDurations := h.k.GetLockableDurations(ctx)
	for _, lockableDuration := range lockableDurations {
		if lockDuration < lockableDuration {
			continue
		}
		for _, coin := range amount {
			currentReward := h.k.GetCurrentReward(coin.GetDenom(), lockableDuration)
			currentReward.IsNewEpoch = true
		}
	}
}

func (h Hooks) OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
}

////////////////////////////  STH END //////////////////////////////////
