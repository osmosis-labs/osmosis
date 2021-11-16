package keeper

func (k Keeper) SuperfluidDelegate(lockID uint64, valAddr string) {
	// TODO: Register a synthetic lockup for superfluid staking
	// synthetic suffix = `superdelegation{valAddr}`
	// TODO: mint OSMO token based on TWAP of locked denom to denom module account
	// TODO: make delegation from module account to the validator
}

func (k Keeper) SuperfluidRedelegate(lockID uint64, newValAddr string) {
	// TODO: Delete previous synthetic lockup, should use native lockup module only or just record lockID - shadow pair on superfluid module?
	// Since synthetic lockup could be used in several places, would be better to create matching on own storage
	// TODO: Create unbonding synthetic lockup for previous shadow
	// synthetic suffix = `redelegating{valAddr}`
	// TODO: Register a synthetic lockup, call SuperfluidDelegate?
	// TODO: Unbonding amount should be modified for TWAP change or not?
}

func (k Keeper) SuperfluidUndelegate(lockID uint64) {
	// Create unbonding synthetic lockup
	// TODO: Unbonding amount should be modified for TWAP change or not?
	// synthetic suffix = `unbonding{valAddr}`
}

func (k Keeper) SuperfluidWithdraw(lockID uint64) {
	// It looks like LP token will be automatically removed by lockup module
	// TODO: If there's any local storage used by superfluid module for each lockID, just clean it up.
	// TODO: automatically done or manually done?
	// TODO: check synthetic suffix = `unbonding{valAddr}`, lockID is matured and removed already on lockup storage
}

// TODO: Need to (eventually) override the existing staking messages and queries, for undelegating, delegating, rewards, and redelegating, to all be going through all superfluid module.
// Want integrators to be able to use the same staking queries and messages
// Eugen’s point: Only rewards message needs to be updated. Rest of messages are fine
// Queries need to be updated
// We can do this at the very end though, since it just relates to queries.
