package keeper_test

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"
)

// TestEpochHook tests that the epoch hook is correctly setting the pool IDs for the osmo and atom pools.
// The epoch hook is run after the pools are set and committed in keeper_test.go. All of the pools are initialized in the setup function in keeper_test.go
// and are available in the suite.pools variable. The pools are filtered to only include the pools that
// have osmo or atom as one of the assets. The pools are then filtered again to only include the pools that
// have the highest liquidity. The pools are then checked to see if the pool IDs are correctly set in the
// osmo and atom stores.
func (suite *KeeperTestSuite) TestEpochHook() {
	// All of the pools initialized in the setup function are available in keeper_test.go
	// akash <-> types.OsmosisDenomination
	// juno <-> types.OsmosisDenomination
	// ethereum <-> types.OsmosisDenomination
	// bitcoin <-> types.OsmosisDenomination
	// canto <-> types.OsmosisDenomination
	// and so on...
	expectedToSee := make(map[string]Pool)
	for _, pool := range suite.pools {

		// Module currently limited to two asset pools
		// Instantiate asset and amounts for the pool
		if len(pool.PoolAssets) == 2 {
			pool.Asset1 = pool.PoolAssets[0].Token.Denom
			pool.Amount1 = pool.PoolAssets[0].Token.Amount
			pool.Asset2 = pool.PoolAssets[1].Token.Denom
			pool.Amount2 = pool.PoolAssets[1].Token.Amount
		}

		if pool.Asset1 == types.OsmosisDenomination || pool.Asset2 == types.OsmosisDenomination || pool.Asset1 == types.AtomDenomination || pool.Asset2 == types.AtomDenomination {
			// create a key that is a combination of asset1 and asset2 in alphabetical order
			key := fmt.Sprintf("%s-%s", pool.Asset1, pool.Asset2)
			if pool.Asset1 > pool.Asset2 {
				key = fmt.Sprintf("%s-%s", pool.Asset2, pool.Asset1)
			}

			if storedPool, ok := expectedToSee[key]; !ok {
				expectedToSee[key] = pool
			} else {
				liquidity := pool.Amount1.Mul(pool.Amount2)
				if liquidity.GT(storedPool.Amount1.Mul(storedPool.Amount2)) {
					expectedToSee[key] = pool
				}
			}
		}
	}

	// The epoch hook is run after the pools are set and committed so all that must be done is the stores must be checked if they are correctly set
	for _, pool := range expectedToSee {
		foundEitherOne := false
		// Check if there is a match with osmo
		if otherDenom, match := types.CheckOsmoAtomDenomMatch(pool.Asset1, pool.Asset2, types.OsmosisDenomination); match {
			poolId, err := suite.App.AppKeepers.ProtoRevKeeper.GetOsmoPool(suite.Ctx, otherDenom)

			// pool ID must exist
			suite.Require().NoError(err)
			suite.Require().Equal(pool.PoolId, poolId)

			foundEitherOne = true
		}

		// Check if there is a match with atom
		if otherDenom, match := types.CheckOsmoAtomDenomMatch(pool.Asset1, pool.Asset2, types.AtomDenomination); match {
			poolId, err := suite.App.AppKeepers.ProtoRevKeeper.GetAtomPool(suite.Ctx, otherDenom)

			// pool ID must exist
			suite.Require().NoError(err)
			suite.Require().Equal(poolId, poolId)

			foundEitherOne = true
		}

		suite.Require().True(foundEitherOne)
	}
}
