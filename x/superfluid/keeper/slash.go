package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

func (k Keeper) SlashLockupsForUnbondingDelegationSlash(ctx sdk.Context, delAddrStr string, valAddrStr string, slashFactor sdk.Dec) {
	delAddr, err := sdk.AccAddressFromBech32(delAddrStr)
	if err != nil {
		panic(err)
	}
	acc := k.GetIntermediaryAccount(ctx, delAddr)
	if acc.Denom == "" { // if delAddr is not intermediary account, pass
		return
	}

	locks := k.lk.GetLocksLongerThanDurationDenom(ctx, acc.Denom+unstakingSuffix(acc.ValAddr), time.Hour*24*21)
	for _, lock := range locks {
		// Only single token lock is allowed here
		slashAmt := lock.Coins[0].Amount.ToDec().Mul(slashFactor).TruncateInt()
		_, err = k.lk.SlashTokensFromLockByID(ctx, lock.ID, sdk.Coins{sdk.NewCoin(lock.Coins[0].Denom, slashAmt)})
		if err != nil {
			panic(err)
		}
	}
}

// Note: Based on sdk.staking.Slash function review, slashed tokens are burnt not sent to community pool
func (k Keeper) SlashLockupsForSlashedOnDelegation(ctx sdk.Context) {
	accs := k.GetAllIntermediaryAccounts(ctx)
	for _, acc := range accs {
		mAddr := acc.GetAddress()
		valAddress, err := sdk.ValAddressFromBech32(acc.ValAddr)
		if err != nil {
			panic(err)
		}

		validator, found := k.sk.GetValidator(ctx, valAddress)
		if !found {
			panic("validator not found")
		}

		// get delegation from intermediary account to the validator
		delegation, found := k.sk.GetDelegation(ctx, mAddr, valAddress)
		if !found {
			continue
		}

		delegatedTokens := validator.TokensFromShares(delegation.Shares).TruncateInt()

		twap := k.GetLastEpochOsmoEquivalentTWAP(ctx, acc.Denom)
		if twap.EpochTwapPrice.IsZero() {
			continue
		}

		// mint OSMO token based on TWAP of locked denom to denom module account
		// Get total delegation from synthetic lockups
		queryCondition := lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         acc.Denom + stakingSuffix(acc.ValAddr),
			Duration:      time.Hour * 24 * 14,
		}
		totalSuperfluidDelegation := k.lk.GetPeriodLocksAccumulation(ctx, queryCondition)
		decAmt := twap.EpochTwapPrice.Mul(totalSuperfluidDelegation.ToDec())
		asset := k.GetSuperfluidAsset(ctx, acc.Denom)
		amt := k.GetRiskAdjustedOsmoValue(ctx, asset, decAmt.RoundInt())

		if !amt.Equal(delegatedTokens) && amt.IsPositive() {
			// (1 - amt/delegatedTokens) describes slash factor
			slashFactor := sdk.OneDec().Sub(delegatedTokens.ToDec().Quo(amt.ToDec()))
			locks := k.lk.GetLocksLongerThanDurationDenom(ctx, queryCondition.Denom, queryCondition.Duration)
			for _, lock := range locks {
				// Only single token lock is allowed here
				slashAmt := lock.Coins[0].Amount.ToDec().Mul(slashFactor).TruncateInt()
				_, err = k.lk.SlashTokensFromLockByID(ctx, lock.ID, sdk.Coins{sdk.NewCoin(lock.Coins[0].Denom, slashAmt)})
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
