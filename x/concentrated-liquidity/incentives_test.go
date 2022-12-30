package concentrated_liquidity_test

// Tests:
// 1. Internal accumulator works
// 2. External accumulator works
// 3. Proper separation between internal and external accums (e.g. can do same token w/ diff rates in each – unintended behavior but good test)
// 4. Sumtree:
// 	  * Get total liquidity >=10s uptime (jointime <= curTime - 10)
// 	  * Get total shares >=10s uptime (jointime <= curTime - 10)
// 	  * Get rewards for a specific position
// 	  * Give reward to LPs >=10s uptime (jointime <= curTime - 10)
// 	  * Claim rewards for a position w/ valid uptime
// 	  * Claim rewards for a position w/ invalid uptime
func (s *KeeperTestSuite) TestInitIncentives() {
	// Sumtree test: regenerate sumtree using accumulation store and check amount.
	// Rerun to ensure adds, removals etc. work as intended.
	
}