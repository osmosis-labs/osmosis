package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
}
