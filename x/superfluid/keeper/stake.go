package keeper

// func name: SuperfluidDelegate(lockID, validator), Register a synthetic lockup for superfluid staking, this creates shadow lockup, mint OSMO token based on TWAP of locked denom to denom module account, make delegation from module account to the validator
// func name: SuperfluidRedelegate(lockID, validator2), Register a synthetic lockup, Create unbonding synthetic lockup
// func name: SuperfluidUndelegate(lockID), Create unbonding synthetic lockup
// func name: SuperfluidWithdraw(lockID) - automatically done or manually done.
// Need to (eventually) override the existing staking messages and queries, for undelegating, delegating, rewards, and redelegating, to all be going through all superfluid module.
// Want integrators to be able to use the same staking queries and messages
// Eugen’s point: Only rewards message needs to be updated. Rest of messages are fine
// Queries need to be updated
// We can do this at the very end though, since it just relates to queries.
