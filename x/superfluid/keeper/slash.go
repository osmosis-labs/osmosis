package keeper

// Add hook for slash event, send slashed LP token amount to community pool, update LP token amount
// Methods to look at:
// SlashUnbondingDelegation
// SlashRedelegation
// burnBondedTokens
// burnUnbondedTokens
// Need to add hooks here, to ensure that instead of sending Osmo to community pool,
// if the osmo is from the superfluid module, we instead burn the osmo, and send equivalent LP shares to community pool
