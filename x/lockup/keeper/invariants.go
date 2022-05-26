package keeper

// DONTCOVER

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

// RegisterInvariants registers all governance invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, keeper Keeper) {
	ir.RegisterRoute(types.ModuleName, "synthetic-lockup-invariant", SyntheticLockupInvariant(keeper))
	ir.RegisterRoute(types.ModuleName, "accumulation-store-invariant", AccumulationStoreInvariant(keeper))
}

func SyntheticLockupInvariant(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		synthlocks := keeper.GetAllSyntheticLockups(ctx)
		for _, synthlock := range synthlocks {
			baselock, err := keeper.GetLockByID(ctx, synthlock.UnderlyingLockId)
			if err != nil {
				panic(err)
			}
			if baselock.ID != synthlock.UnderlyingLockId {
				return sdk.FormatInvariant(types.ModuleName, "synthetic-lockup-invariant",
					fmt.Sprintf("\tSynthetic lock denom %s\n\tUnderlying lock ID: %d\n\tActual underying lock ID: %d\n",
						synthlock.SynthDenom, synthlock.UnderlyingLockId, baselock.ID,
					)), true
			}
		}

		return sdk.FormatInvariant(types.ModuleName, "synthetic-lockup-invariant", "All synthetic lockup invariant passed"), false
	}
}

// AccumulationStoreInvariant ensures that the sum of all lockups at a given duration
// is equal to the value stored within the accumulation store.
func AccumulationStoreInvariant(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		moduleAcc := keeper.ak.GetModuleAccount(ctx, types.ModuleName)
		balances := keeper.bk.GetAllBalances(ctx, moduleAcc.GetAddress())

		// check 1s, 1 day, 1 week, 2 weeks
		durations := []time.Duration{
			time.Second,
			time.Hour * 24,
			time.Hour * 24 * 7,
			time.Hour * 24 * 14,
		}

		// loop all denoms on lockup module
		for _, coin := range balances {
			denom := coin.Denom
			for _, duration := range durations {
				accumulation := keeper.GetPeriodLocksAccumulation(ctx, types.QueryCondition{
					LockQueryType: types.ByDuration,
					Denom:         denom,
					Duration:      duration,
				})

				locks := keeper.GetLocksLongerThanDurationDenom(ctx, denom, duration)
				lockupSum := sdk.ZeroInt()
				for _, lock := range locks {
					lockupSum = lockupSum.Add(lock.Coins.AmountOf(denom))
				}

				if !accumulation.Equal(lockupSum) {
					return sdk.FormatInvariant(types.ModuleName, "accumulation-store-invariant",
						fmt.Sprintf("\taccumulation store value does not fit actual lockup sum: %s != %s\n",
							accumulation.String(), lockupSum.String(),
						)), true
				}
			}
		}

		return sdk.FormatInvariant(types.ModuleName, "accumulation-store-invariant", "All lockup accumulation invariant passed"), false
	}
}
