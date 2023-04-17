package keeper_test

import (
	"fmt"
	"strings"

	"testing"

	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// BenchmarkEpochHook benchmarks the epoch hook. In particular, it benchmarks the UpdatePools function.
func BenchmarkEpochHook(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	// Setup the test suite
	suite := new(KeeperTestSuite)
	suite.SetT(&testing.T{})
	suite.SetupTest()

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		suite.App.ProtoRevKeeper.UpdatePools(suite.Ctx)
		b.StopTimer()
	}
}

// TestEpochHook tests that the epoch hook is correctly setting the pool IDs for all base denoms. Base denoms are the denoms that will
// be used for cyclic arbitrage and must be stored in the keeper. The epoch hook is run after the pools are set and committed in keeper_test.go.
// All of the pools are initialized in the setup function in keeper_test.go and are available in the suite.pools variable. In this test
// function, the pools are filtered to only include the pools that have at least one base denom as an asset. The pools are then filtered
// again to only include the pools that have the highest liquidity. The pools are then checked to see if the pool IDs are correctly set in the
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
	baseDenoms, err := suite.App.ProtoRevKeeper.GetAllBaseDenoms(suite.Ctx)
	suite.Require().NoError(err)
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

	// Iterate and ensure that the keeper has the correct pool IDs for the base denoms
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

		// In the case where the pool contains two base denoms, make sure that they both store the same pool ID
		if contains(baseDenoms, denoms[0]) && contains(baseDenoms, denoms[1]) {
			poolId, err := suite.App.ProtoRevKeeper.GetPoolForDenomPair(suite.Ctx, denoms[0], denoms[1])
			suite.Require().NoError(err)
			suite.Require().Equal(pool.PoolId, poolId)

			otherPoolId, err := suite.App.ProtoRevKeeper.GetPoolForDenomPair(suite.Ctx, denoms[1], denoms[0])
			suite.Require().NoError(err)
			suite.Require().Equal(pool.PoolId, otherPoolId)

			suite.Require().Equal(poolId, otherPoolId)
		}

		if poolVisited {
			totalActuallySeen++
		}
	}

	suite.Require().Equal(totalNumberExpected, totalActuallySeen)
}

func contains(baseDenoms []types.BaseDenom, denomToMatch string) bool {
	for _, baseDenom := range baseDenoms {
		if baseDenom.Denom == denomToMatch {
			return true
		}
	}
	return false
}
