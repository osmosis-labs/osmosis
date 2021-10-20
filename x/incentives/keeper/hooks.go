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
			err := k.F1Distribute(ctx, &gauge)
			if err != nil {
				panic(err)
			}
			if !gauge.IsPerpetual && gauge.NumEpochsPaidOver <= gauge.FilledEpochs {
				if err := k.FinishDistribution(ctx, gauge); err != nil {
					panic(err)
				}
			}
		}

		k.hooks.AfterEpochDistribution(ctx)
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

var _ lockuptypes.LockupHooks = Hooks{}

func (h Hooks) OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	lockRef, err := h.k.lk.GetLockByID(ctx, lockID)
	if err != nil || lockRef == nil {
		return
	}
	prevLock := *lockRef

	// coin of amount 0 needs to be tracked but sdk.Coins.Sub sanitize coin of amount 0.
	prevLockCoins := []sdk.Coin{}
	for _, lockedCoin := range prevLock.Coins {
		prevLockCoins = append(prevLockCoins, sdk.NewCoin(lockedCoin.Denom, lockedCoin.Amount.Sub(amount.AmountOf(lockedCoin.Denom))))
	}

	prevLock.Coins = prevLockCoins

	lockReward, err := h.k.GetPeriodLockReward(ctx, lockID)
	if err != nil {
		panic(err)
	}
	epochInfo := h.k.GetEpochInfo(ctx)
	lockableDurations := h.k.GetLockableDurations(ctx)
	err = h.k.UpdateHistoricalReward(ctx, amount, lockDuration, epochInfo, lockableDurations)
	if err != nil {
		panic(err)
	}
	err = h.k.UpdatePeriodLockReward(ctx, prevLock, lockReward, epochInfo, lockableDurations)
	if err != nil {
		panic(err)
	}
}

func (h Hooks) OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	lock, err := h.k.lk.GetLockByID(ctx, lockID)
	if err != nil {
		panic(err)
	}
	lockReward, err := h.k.GetPeriodLockReward(ctx, lockID)
	if err != nil {
		panic(err)
	}
	epochInfo := h.k.GetEpochInfo(ctx)
	lockableDurations := h.k.GetLockableDurations(ctx)
	if lock.Coins.IsAnyGT(amount) {
		err = h.k.UpdateHistoricalReward(ctx, amount, lockDuration, epochInfo, lockableDurations)
		if err != nil {
			panic(err)
		}
		err = h.k.UpdatePeriodLockReward(ctx, *lock, lockReward, epochInfo, lockableDurations)
		if err != nil {
			panic(err)
		}
	} else {
		owner, err := sdk.AccAddressFromBech32(lock.Owner)
		if err != nil {
			panic(err)
		}
		_, err = h.k.ClaimLockReward(ctx, *lock, owner)
		if err != nil {
			panic(err)
		}
		h.k.clearPeriodLockReward(ctx, lockID)
	}
}
