package keeper_test

import (
	"fmt"
	"strings"
)

// TestEpochHook tests that the epoch hook is correctly setting the pool IDs for all base denoms.
// The epoch hook is run after the pools are set and committed in keeper_test.go. All of the pools are initialized in
// the setup function in keeper_test.go and are available in the suite.pools variable. In this test function, the pools are filtered to
// only include the pools that have a base denom as one of the assets. Base denoms are the denoms that will be used for cyclic arbitrage
// and must be stored in the keeper. The pools are then filtered again to only include the pools that
// have the highest liquidity. The pools are then checked to see if the pool IDs are correctly set in the
// DenomPairToPool stores.
func (suite *KeeperTestSuite) TestEpochHook() {
	// All of the pools initialized in the setup function are available in keeper_test.go
	// akash <-> types.OsmosisDenomination
	// juno <-> types.OsmosisDenomination
	// ethereum <-> types.OsmosisDenomination
	// bitcoin <-> types.OsmosisDenomination
	// canto <-> types.OsmosisDenomination
	// and so on...

	totalNumberExpected := 0
	expectedToSee := make(map[string]Pool)
	baseDenoms := suite.App.ProtoRevKeeper.GetAllBaseDenoms(suite.Ctx)
	for _, pool := range suite.pools {

		// Module currently limited to two asset pools
		// Instantiate asset and amounts for the pool
		if len(pool.PoolAssets) == 2 {
			pool.Asset1 = pool.PoolAssets[0].Token.Denom
			pool.Amount1 = pool.PoolAssets[0].Token.Amount
			pool.Asset2 = pool.PoolAssets[1].Token.Denom
			pool.Amount2 = pool.PoolAssets[1].Token.Amount
		}

		if contains(baseDenoms, pool.Asset1) || contains(baseDenoms, pool.Asset2) {
			// create a key that is a combination of asset1 and asset2 in alphabetical order
			key := fmt.Sprintf("%s-%s", pool.Asset1, pool.Asset2)
			if pool.Asset1 > pool.Asset2 {
				key = fmt.Sprintf("%s-%s", pool.Asset2, pool.Asset1)
			}

			if storedPool, ok := expectedToSee[key]; !ok {
				expectedToSee[key] = pool
				totalNumberExpected++
			} else {
				liquidity := pool.Amount1.Mul(pool.Amount2)
				if liquidity.GT(storedPool.Amount1.Mul(storedPool.Amount2)) {
					expectedToSee[key] = pool
				}
			}
		}
	}

	// The epoch hook is run after the pools are set and committed so all that must be done is the stores must be checked if they are correctly set
	totalActuallySeen := 0
	for key, pool := range expectedToSee {
		poolVisited := false

		// split the key and check if it contains a base denom
		denoms := strings.Split(key, "-")
		if contains(baseDenoms, denoms[0]) {
			poolId, err := suite.App.ProtoRevKeeper.GetPoolForDenomPair(suite.Ctx, denoms[0], denoms[1])
			suite.Require().NoError(err)
			suite.Require().Equal(pool.PoolId, poolId)
			poolVisited = true
		}

		if contains(baseDenoms, denoms[1]) {
			poolId, err := suite.App.ProtoRevKeeper.GetPoolForDenomPair(suite.Ctx, denoms[1], denoms[0])
			suite.Require().NoError(err)
			suite.Require().Equal(pool.PoolId, poolId)
			poolVisited = true
		}

		if poolVisited {
			totalActuallySeen++
		}
	}

	suite.Require().Equal(totalNumberExpected, totalActuallySeen)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
