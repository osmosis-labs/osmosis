package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

var defaultPoolId = uint64(1)

func (s *KeeperTestSuite) TestCreateAndGetUptimeAccumulators() {
	// We expect there to be len(types.SupportedUptimes) number of initialized accumulators
	// for a successful pool creation. We calculate this upfront to ensure test compatibility
	// if the uptimes we support ever change.
	curExpectedAccumValues := []sdk.DecCoins{}
	for range types.SupportedUptimes {
		curExpectedAccumValues = append(curExpectedAccumValues, cl.EmptyCoins)
	}
	s.Require().Equal(len(types.SupportedUptimes), len(curExpectedAccumValues))

	type initUptimeAccumTest struct {
		name                string
		poolId              uint64
		initializePool      bool
		expectedAccumValues []sdk.DecCoins

		expectedPass bool
	}
	tests := []initUptimeAccumTest{
		{
			name:                "default pool setup",
			poolId:              defaultPoolId,
			initializePool:      true,
			expectedAccumValues: curExpectedAccumValues,
			expectedPass:        true,
		},
		{
			name:                "setup with different poolId",
			poolId:              defaultPoolId + 1,
			initializePool:      true,
			expectedAccumValues: curExpectedAccumValues,
			expectedPass:        true,
		},
		{
			name:                "pool not initialized",
			initializePool:      false,
			poolId:              defaultPoolId,
			expectedAccumValues: []sdk.DecCoins{},
			expectedPass:        false,
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()
			clKeeper := s.App.ConcentratedLiquidityKeeper

			// system under test
			var err error
			if tc.initializePool {
				err = clKeeper.CreateUptimeAccumulators(s.Ctx, tc.poolId)
			}
			poolUptimeAccumulators, err := clKeeper.GetUptimeAccumulators(s.Ctx, tc.poolId)

			if tc.expectedPass {
				s.Require().NoError(err)

				// ensure number of uptime accumulators match supported uptimes
				s.Require().Equal(len(tc.expectedAccumValues), len(poolUptimeAccumulators))

				// ensure that each uptime was initialized to the correct value (sdk.DecCoins(nil))
				accumValues := []sdk.DecCoins{}
				for _, accum := range poolUptimeAccumulators {
					accumValues = append(accumValues, accum.GetValue())
				}
				s.Require().Equal(tc.expectedAccumValues, accumValues)
			} else {
				s.Require().Error(err)

				// ensure no accumulators exist for an uninitialized pool
				s.Require().Equal(0, len(poolUptimeAccumulators))
			}
		})
	}
}
