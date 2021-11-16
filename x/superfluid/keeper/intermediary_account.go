package keeper

// Minted OSMO amount
// LP token denom
// LP token amount
// Unique AccAddress derivated from LP token denom: Can use NewEmptyModuleAccount(LP token denom)
// Delegation amount (This can be just fetched from AccAddress using keeper)
// Slashed amount = Minted OSMO amount - Delegation amount
// On TWAP change, set target OSMO mint amount to TWAP * LP token amount
