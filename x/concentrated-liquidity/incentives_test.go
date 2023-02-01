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
		poolId              uint64
		initializePoolAccum bool
		expectedAccumValues []sdk.DecCoins

		expectedPass bool
	}
	tests := map[string]initUptimeAccumTest{
		"default pool setup": {
			poolId:              defaultPoolId,
			initializePoolAccum: true,
			expectedAccumValues: curExpectedAccumValues,
			expectedPass:        true,
		},
		"setup with different poolId": {
			poolId:              defaultPoolId + 1,
			initializePoolAccum: true,
			expectedAccumValues: curExpectedAccumValues,
			expectedPass:        true,
		},
		"pool not initialized": {
			initializePoolAccum: false,
			poolId:              defaultPoolId,
			expectedAccumValues: []sdk.DecCoins{},
			expectedPass:        false,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			clKeeper := s.App.ConcentratedLiquidityKeeper

			// system under test
			if tc.initializePoolAccum {
				err := clKeeper.CreateUptimeAccumulators(s.Ctx, tc.poolId)
				s.Require().NoError(err)
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

func (s *KeeperTestSuite) TestGetUptimeAccumulatorName() {
	type getUptimeNameTest struct {
		poolId            uint64
		uptimeId          uint64
		expectedAccumName string
	}
	tests := map[string]getUptimeNameTest{
		"pool id 1, uptime id 0": {
			poolId:            defaultPoolId,
			uptimeId:          uint64(0),
			expectedAccumName: "uptime/1/0",
		},
		"pool id 1, uptime id 999": {
			poolId:            defaultPoolId,
			uptimeId:          uint64(999),
			expectedAccumName: "uptime/1/999",
		},
		"pool id 999, uptime id 1": {
			poolId:            uint64(999),
			uptimeId:          uint64(1),
			expectedAccumName: "uptime/999/1",
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()

			// system under test
			accumName := cl.GetUptimeAccumulatorName(tc.poolId, tc.uptimeId)
			s.Require().Equal(tc.expectedAccumName, accumName)
		})
	}
}

func (s *KeeperTestSuite) TestCreateAndGetUptimeAccumulatorValues() {
	// We expect there to be len(types.SupportedUptimes) number of initialized accumulators
	// for a successful pool creation. 
	// We re-calculate these values each time to ensure test compatibility if the uptimes 
	// we support ever change.
	curExpectedAccumValues := []sdk.DecCoins{}
	hundredTokensSingleDenom := []sdk.DecCoins{}
	hundredTokensMultiDenom := []sdk.DecCoins{}
	twoHundredTokensMultiDenom := []sdk.DecCoins{}
	varyingTokensSingleDenom := []sdk.DecCoins{}
	varyingTokensMultiDenom := []sdk.DecCoins{}
	for i := range types.SupportedUptimes {
		curExpectedAccumValues = append(curExpectedAccumValues, cl.EmptyCoins)
		hundredTokensSingleDenom = append(hundredTokensSingleDenom, sdk.NewDecCoins(cl.HundredFooCoins))
		hundredTokensMultiDenom = append(hundredTokensMultiDenom, sdk.NewDecCoins(cl.HundredFooCoins, cl.HundredBarCoins))
		twoHundredTokensMultiDenom = append(twoHundredTokensMultiDenom, sdk.NewDecCoins(cl.HundredFooCoins.Add(cl.HundredFooCoins), cl.HundredBarCoins.Add(cl.HundredBarCoins)))
		varyingTokensSingleDenom = append(varyingTokensSingleDenom, sdk.NewDecCoins(cl.HundredFooCoins.Add(sdk.NewDecCoin("foo", sdk.NewInt(int64(i))))))
		varyingTokensMultiDenom = append(varyingTokensMultiDenom, sdk.NewDecCoins(cl.HundredFooCoins.Add(sdk.NewDecCoin("foo", sdk.NewInt(int64(i)))), cl.HundredBarCoins.Add(sdk.NewDecCoin("bar", sdk.NewInt(int64(i * 3))))))
	}
	s.Require().Equal(len(types.SupportedUptimes), len(curExpectedAccumValues))
	s.Require().Equal(len(types.SupportedUptimes), len(hundredTokensSingleDenom))
	s.Require().Equal(len(types.SupportedUptimes), len(hundredTokensMultiDenom))
	s.Require().Equal(len(types.SupportedUptimes), len(twoHundredTokensMultiDenom))
	s.Require().Equal(len(types.SupportedUptimes), len(varyingTokensSingleDenom))
	s.Require().Equal(len(types.SupportedUptimes), len(varyingTokensMultiDenom))

	type initUptimeAccumTest struct {
		poolId              uint64
		initializePoolAccums      bool
		addedAccumValues 	[]sdk.DecCoins
		numTimesAdded		int
		expectedAccumValues []sdk.DecCoins

		expectedPass bool
	}
	tests := map[string]initUptimeAccumTest{
		"hundred of a single denom in each accumulator added once": {
			poolId:              defaultPoolId,
			initializePoolAccums:      true,
			addedAccumValues:	hundredTokensSingleDenom,
			numTimesAdded: 1,
			expectedAccumValues: hundredTokensSingleDenom,
			expectedPass:        true,
		},
		"hundred of multiple denom in each accumulator added once": {
			poolId:              defaultPoolId,
			initializePoolAccums:      true,
			addedAccumValues:	hundredTokensMultiDenom,
			numTimesAdded: 1,
			expectedAccumValues: hundredTokensMultiDenom,
			expectedPass:        true,
		},
		"varying amounts of single denom in each accumulator added once": {
			poolId:              defaultPoolId,
			initializePoolAccums:      true,
			addedAccumValues:	varyingTokensSingleDenom,
			numTimesAdded: 1,
			expectedAccumValues: varyingTokensSingleDenom,
			expectedPass:        true,
		},
		"varying of multiple denoms in each accumulator added once": {
			poolId:              defaultPoolId,
			initializePoolAccums:      true,
			addedAccumValues:	varyingTokensMultiDenom,
			numTimesAdded: 1,
			expectedAccumValues: varyingTokensMultiDenom,
			expectedPass:        true,
		},
		"hundred of multiple denom in each accumulator added twice": {
			poolId:              defaultPoolId,
			initializePoolAccums:      true,
			addedAccumValues:	hundredTokensMultiDenom,
			numTimesAdded: 2,
			expectedAccumValues: twoHundredTokensMultiDenom,
			expectedPass:        true,
		},
		"setup with different poolId": {
			poolId:              defaultPoolId + 1,
			initializePoolAccums:      true,
			addedAccumValues:	hundredTokensSingleDenom,
			numTimesAdded: 1,
			expectedAccumValues: hundredTokensSingleDenom,
			expectedPass:        true,
		},
		"pool not initialized": {
			initializePoolAccums:      false,
			poolId:              defaultPoolId,
			addedAccumValues:	hundredTokensSingleDenom,
			numTimesAdded: 1,
			expectedAccumValues: []sdk.DecCoins{},
			expectedPass:        false,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			clKeeper := s.App.ConcentratedLiquidityKeeper

			// system under test
			var err error
			if tc.initializePoolAccums {
				err = clKeeper.CreateUptimeAccumulators(s.Ctx, tc.poolId)
				s.Require().NoError(err)

				poolUptimeAccumulators, err := clKeeper.GetUptimeAccumulators(s.Ctx, tc.poolId)
				s.Require().NoError(err)
				
				for i := 0; i < tc.numTimesAdded; i++ {
					for uptimeId, uptimeAccum := range poolUptimeAccumulators {
						uptimeAccum.AddToAccumulator(tc.addedAccumValues[uptimeId])
					}
					poolUptimeAccumulators, err = clKeeper.GetUptimeAccumulators(s.Ctx, tc.poolId)
					s.Require().NoError(err)
				}
			}
			poolUptimeAccumulatorValues, err := clKeeper.GetUptimeAccumulatorValues(s.Ctx, tc.poolId)

			if tc.expectedPass {
				s.Require().NoError(err)

				// ensure number of uptime accumulators match supported uptimes
				s.Require().Equal(len(tc.expectedAccumValues), len(poolUptimeAccumulatorValues))

				// ensure that each uptime was initialized to the correct value (sdk.DecCoins(nil))
				s.Require().Equal(tc.expectedAccumValues, poolUptimeAccumulatorValues)
			} else {
				s.Require().Error(err)

				// ensure no accumulators exist for an uninitialized pool
				s.Require().Equal(0, len(poolUptimeAccumulatorValues))
			}
		})
	}
}
