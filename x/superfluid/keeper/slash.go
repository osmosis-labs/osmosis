package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (k Keeper) SlashLockupsForUnbondingDelegationSlash(ctx sdk.Context, delAddrStr string, valAddrStr string, slashFactor sdk.Dec) {
	delAddr, err := sdk.AccAddressFromBech32(delAddrStr)
	if err != nil {
		panic(err)
	}

	// TODO: What?? The intermediary accounts aren't serialized by delAddr??
	// Lets replace with the following pseudocode
	// locks := lk.GetAllSyntheticLockupsByAddr(delAddr)
	// for _, lock := range locks {
	// 	denom, val := parse_denom(lock.suffix)
	//  // No need for infraction height checks, as this delegator is already confirmed as needing a slash.
	// 	if val == relevant_val {
	// 	   do_slash(lock)
	// 	}
	//  }
	acc := k.GetIntermediaryAccount(ctx, delAddr)
	// if delAddr is not intermediary account, pass
	if acc.Denom == "" {
		return
	}

	// Get lockups longer or equal to SuperfluidUnbondDuration
	params := k.GetParams(ctx)
	locks := k.lk.GetLocksLongerThanDurationDenom(ctx, acc.Denom+unstakingSuffix(acc.ValAddr), params.UnbondingDuration)
	for _, lock := range locks {
		// slashing only applies to synthetic lockup amount
		synthLock, err := k.lk.GetSyntheticLockup(ctx, lock.ID, unstakingSuffix(acc.ValAddr))
		if err != nil {
			k.Logger(ctx).Error(err.Error())
			continue
		}

		// Only single token lock is allowed here
		slashAmt := synthLock.Coins[0].Amount.ToDec().Mul(slashFactor).TruncateInt()
		osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			_, err = k.lk.SlashTokensFromLockByID(cacheCtx, lock.ID, sdk.Coins{sdk.NewCoin(lock.Coins[0].Denom, slashAmt)})
			return err
		})
	}
}

// Note: Based on sdk.staking.Slash function review, slashed tokens are burnt not sent to community pool
func (k Keeper) SlashLockupsForValidatorSlash(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	accs := k.GetAllIntermediaryAccounts(ctx)
	valAccs := []types.SuperfluidIntermediaryAccount{}
	for _, acc := range accs {
		if acc.ValAddr != valAddr.String() { // only apply for slashed validator
			continue
		}
		valAccs = append(valAccs, acc)
	}

	for _, acc := range valAccs {
		// mint OSMO token based on TWAP of locked denom to denom module account
		// Get total delegation from synthetic lockups
		queryCondition := lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         acc.Denom + stakingSuffix(acc.ValAddr),
			Duration:      time.Hour * 24 * 14,
		}

		// (1 - amt/delegatedTokens) describes slash factor
		locks := k.lk.GetLocksLongerThanDurationDenom(ctx, queryCondition.Denom, queryCondition.Duration)
		for _, lock := range locks {
			// slashing only applies to synthetic lockup amount
			synthLock, err := k.lk.GetSyntheticLockup(ctx, lock.ID, stakingSuffix(acc.ValAddr))
			if err != nil {
				k.Logger(ctx).Error(err.Error())
				continue
			}

			// Only single token lock is allowed here
			slashAmt := synthLock.Coins[0].Amount.ToDec().Mul(fraction).TruncateInt()
			osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
				_, err := k.lk.SlashTokensFromLockByID(cacheCtx, lock.ID, sdk.Coins{sdk.NewCoin(lock.Coins[0].Denom, slashAmt)})
				return err
			})
		}
	}
}
