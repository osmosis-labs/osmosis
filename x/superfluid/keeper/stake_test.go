package keeper_test

func (suite *KeeperTestSuite) TestSuperfluidDelegate() {
	// TODO: create validator
	// TODO: register a LP token as a superfluid asset
	// TODO: set OSMO TWAP price for LP token
	// TODO: create lockup of LP token
	// TODO: call SuperfluidDelegate
	// TODO: Check superfluid delegate result error
	// TODO: Check synthetic lockup creation
	// TODO: Check intermediary account creation
	// TODO: Check gauge creation
	// TODO: Check lockID connection with intermediary account
	// TODO: Check delegation from intermediary account to validator
	// TODO: add table driven test for all edge cases
}

func (suite *KeeperTestSuite) TestSuperfluidUndelegate() {
	// TODO: do SuperfluidDelgate to test undelegation - utility function
	// TODO: add test for SuperfluidUndelegate
	// TODO: Check superfluid delegate result error
	// TODO: check synthetic lockup deletion for delegation
	// TODO: check unbonding synthetic lockup creation
}

func (suite *KeeperTestSuite) TestSuperfluidRedelegate() {
	// TODO: do SuperfluidDelgate to test undelegation - utility function
	// TODO: add test for SuperfluidRedelegate
	// TODO: check the changes for undelegate function call changes
	// TODO: check the changes for new delegation function call changes
}

func (suite *KeeperTestSuite) TestRefreshIntermediaryDelegationAmounts() {
	// TODO: add test for refreshIntermediaryDelegationAmounts
}
