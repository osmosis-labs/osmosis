package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (k Keeper) SlashLockupsForUnbondingDelegationSlash(ctx sdk.Context, delAddrStr string, valAddrStr string, slashFactor sdk.Dec) {
	delAddr, err := sdk.AccAddressFromBech32(delAddrStr)
	if err != nil {
		panic(err)
	}

	acc := k.GetIntermediaryAccount(ctx, delAddr)
	// if delAddr is not intermediary account, pass
	if acc.Denom == "" {
		return
	}

	// Get lockups longer or equal to SuperfluidUnbondDuration
	params := k.GetParams(ctx)
	locks := k.lk.GetLocksLongerThanDurationDenom(ctx, acc.Denom+unstakingSuffix(acc.ValAddr), params.UnbondingDuration)
	for _, lock := range locks {
		// Only single token lock is allowed here
		slashAmt := lock.Coins[0].Amount.ToDec().Mul(slashFactor).TruncateInt()
		cacheCtx, write := ctx.CacheContext()
		_, err = k.lk.SlashTokensFromLockByID(cacheCtx, lock.ID, sdk.Coins{sdk.NewCoin(lock.Coins[0].Denom, slashAmt)})
		if err != nil {
			k.Logger(ctx).Error(err.Error())
		} else {
			write()
		}
	}
}

// Note: Based on sdk.staking.Slash function review, slashed tokens are burnt not sent to community pool
func (k Keeper) SlashLockupsForValidatorSlash(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	accs := k.GetAllIntermediaryAccounts(ctx)
	for _, acc := range accs {
		if acc.ValAddr != valAddr.String() { // only apply for slashed validator
			continue
		}

		// mint OSMO token based on TWAP of locked denom to denom module account
		// Get total delegation from synthetic lockups
		queryCondition := lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         acc.Denom + stakingSuffix(acc.ValAddr),
			Duration:      types.OsmosisUnbondingPeriod,
		}

		// (1 - amt/delegatedTokens) describes slash factor
		locks := k.lk.GetLocksLongerThanDurationDenom(ctx, queryCondition.Denom, queryCondition.Duration)
		for _, lock := range locks {
			// Only single token lock is allowed here
			slashAmt := lock.Coins[0].Amount.ToDec().Mul(fraction).TruncateInt()
			cacheCtx, write := ctx.CacheContext()
			_, err := k.lk.SlashTokensFromLockByID(cacheCtx, lock.ID, sdk.Coins{sdk.NewCoin(lock.Coins[0].Denom, slashAmt)})
			if err != nil {
				k.Logger(ctx).Error(err.Error())
			} else {
				write()
			}
		}
	}
}
