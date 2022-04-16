package keeper

import (
	"time"

	"github.com/osmosis-labs/osmosis/v8/osmoutils"
	lockuptypes "github.com/osmosis-labs/osmosis/v8/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SlashLockupsForValidatorSlash should be called before the validator at valAddr is slashed.
// This function is responsible for inspecting every intermediate account to valAddr.
// For each intermediate account IA, it slashes every constituent delegation behind IA.
// Furthermore, if the infraction height is sufficiently old, slashes unbondings
// Note: Based on sdk.staking.Slash function review, slashed tokens are burnt not sent to community pool
// we ignore that, and send the underliyng tokens to the community pool anyway.
func (k Keeper) SlashLockupsForValidatorSlash(ctx sdk.Context, valAddr sdk.ValAddress, infractionHeight int64, slashFactor sdk.Dec) {
	// Important note: The SDK slashing for historical heights is wrong.
	// It defines a "slash amount" off of the live staked amount.
	// Then it charges all the unbondings & redelegations at the slash factor.
	// It then creates a new slash factor for the amount remaining to be charged from the slash amount,
	// across all the live accounts.
	// This is the "effectiveSlashFactor".
	//
	// The SDK's design is wack / wrong in our view, and this was a pre Cosmos Hub
	// launch hack that never got remedied.
	// We are not concerned about maximal consistency with the SDK, and instead charge slashFactor to
	// both unbonding and live delegations. Rather than slashFactor to unbonding delegations,
	// and effectiveSlashFactor to new delegations.
	accs := k.GetIntermediaryAccountsForVal(ctx, valAddr)

	// for every intermediary account, we first slash the live tokens comprosing delegated to it,
	// and then all of its unbonding delegations.
	// We do these slashes as burns.
	for _, acc := range accs {
		locks := k.lk.GetLocksLongerThanDurationDenom(ctx, acc.Denom, time.Second)
		for _, lock := range locks {
			// slashing only applies to synthetic lockup amount
			synthLock, err := k.lk.GetSyntheticLockup(ctx, lock.ID, stakingSyntheticDenom(acc.Denom, acc.ValAddr))
			// synth lock doesn't exist for bonding
			if err != nil {
				synthLock, err = k.lk.GetSyntheticLockup(ctx, lock.ID, unstakingSyntheticDenom(acc.Denom, acc.ValAddr))
				// synth lock doesn't exist for unbonding
				// => no superfluid staking on this lock ID, so continue
				if err != nil {
					continue
				}
			}

			// slash the lock whether its bonding or unbonding.
			// this overslashes unbondings that started unbonding before the slash infraction,
			// but this seems to be an acceptable trade-off based upon choices taken in the SDK.
			k.slashSynthLock(ctx, synthLock, slashFactor)
		}
	}
}

func (k Keeper) slashSynthLock(ctx sdk.Context, synthLock *lockuptypes.SyntheticLock, slashFactor sdk.Dec) {
	// Only single token lock is allowed here
	lock, _ := k.lk.GetLockByID(ctx, synthLock.UnderlyingLockId)
	slashAmt := lock.Coins[0].Amount.ToDec().Mul(slashFactor).TruncateInt()
	slashCoins := sdk.NewCoins(sdk.NewCoin(lock.Coins[0].Denom, slashAmt))
	_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		// These tokens get moved to the community pool.
		_, err := k.lk.SlashTokensFromLockByID(cacheCtx, lock.ID, slashCoins)
		return err
	})
}
