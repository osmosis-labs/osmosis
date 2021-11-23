package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

// TODO:
// Add hook for slash event, send slashed LP token amount to community pool, update LP token amount
// Methods to look at:
// SlashUnbondingDelegation
// SlashRedelegation
// burnBondedTokens
// burnUnbondedTokens
// Need to add hooks here, to ensure that instead of sending Osmo to community pool,
// if the osmo is from the superfluid module, we instead burn the osmo, and send equivalent LP shares to community pool

// TODO:
// slashing
// 	Currently for double signs, we iterate over all unbondings and all redelegations. We handle slashing delegated tokens, via a “rebase” factor.
// 	Meaning, that if we have a 10% slash say, we just alter the conversion rate between “delegation pool shares” and “osmo” when withdrawing your stake.
// 	Now in our case, we currently don’t have a method for a “rebase” factor in synthetic lockups.
// 	Eugen: We can add this rebase factor to our Superfluid module, to be executed upon MsgUnbondStake or w/e its called
// 	Dev: I don’t think we need to worry about deferring iteration

func (k Keeper) slashLockupsForSlashedOnDelegation(ctx sdk.Context) {
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

		// undelegate full amount from the validator
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
		decAmt := twap.EpochTwapPrice.Mul(sdk.Dec(totalSuperfluidDelegation))
		asset := k.GetSuperfluidAsset(ctx, acc.Denom)
		amt := k.GetRiskAdjustedOsmoValue(ctx, asset, decAmt.RoundInt())

		if !amt.Equal(delegatedTokens) {
			// (1 - amt/delegatedTokens) describes slash factor
			slashFactor := sdk.OneDec().Sub(delegatedTokens.ToDec().Quo(amt.ToDec()))
			locks := k.lk.GetLocksLongerThanDurationDenom(ctx, queryCondition.Denom, queryCondition.Duration)
			for _, lock := range locks {
				// Only single token lock is allowed here
				slashAmt := lock.Coins[0].Amount.ToDec().Mul(slashFactor).TruncateInt()
				k.lk.SlashTokensFromLockByID(ctx, lock.ID, sdk.Coins{sdk.NewCoin(lock.Coins[0].Denom, slashAmt)})
			}
		}
	}
}
