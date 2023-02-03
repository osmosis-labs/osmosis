package concentrated_liquidity_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

var (
	defaultPoolId = uint64(1)
	defaultLowerTick = -100
	defaultUpperTick = 100

	testAddressOne   = sdk.AccAddress([]byte("addr1_______________"))
	testAddressTwo   = sdk.AccAddress([]byte("addr2_______________"))

	testAccumOne = "testAccumOne"

	testDenomOne = "denomOne"
	testDenomTwo = "denomTwo"
	testDenomThree = "denomThree"

	testAmountOne = sdk.NewDec(2 << 60)
	testAmountTwo = sdk.NewDec(2 << 60)
	testAmountThree = sdk.NewDec(2 << 60)

	testEmissionOne = sdk.MustNewDecFromStr("0.000001")
	testEmissionTwo = sdk.MustNewDecFromStr("0.0783")
	testEmissionThree = sdk.MustNewDecFromStr("165.4")

	defaultBlockTime = time.Now()
	defaultTimeBuffer = time.Hour
	defaultStartTime = defaultBlockTime.Add(defaultTimeBuffer)

	testUptimeOne = types.SupportedUptimes[0]
	testUptimeTwo = types.SupportedUptimes[1]
	testUptimeThree = types.SupportedUptimes[2]

	incentiveRecordOne = types.IncentiveRecord{
		IncentiveDenom: testDenomOne, 
		RemainingAmount: testAmountOne, 
		EmissionRate: testEmissionOne, 
		StartTime: defaultStartTime, 
		MinUptime: testUptimeOne,
	}

	incentiveRecordTwo = types.IncentiveRecord{
		IncentiveDenom: testDenomTwo, 
		RemainingAmount: testAmountTwo, 
		EmissionRate: testEmissionTwo, 
		StartTime: defaultStartTime, 
		MinUptime: testUptimeTwo,
	}

	incentiveRecordThree = types.IncentiveRecord{
		IncentiveDenom: testDenomThree, 
		RemainingAmount: testAmountThree, 
		EmissionRate: testEmissionThree, 
		StartTime: defaultStartTime, 
		MinUptime: testUptimeThree,
	}
)

type ExpectedUptimes struct {
	emptyExpectedAccumValues []sdk.DecCoins
	hundredTokensSingleDenom []sdk.DecCoins
	hundredTokensMultiDenom []sdk.DecCoins
	twoHundredTokensMultiDenom []sdk.DecCoins
	varyingTokensSingleDenom []sdk.DecCoins
	varyingTokensMultiDenom []sdk.DecCoins
}

// getExpectedUptimes returns a base set of expected values for testing based on the number
// of supported uptimes at runtime. This abstraction exists only to ensure backwards-compatibility
// of incentives-related tests if the supported uptimes are ever changed.
func getExpectedUptimes() ExpectedUptimes {
	expUptimes := ExpectedUptimes{
		emptyExpectedAccumValues: []sdk.DecCoins{},
		hundredTokensSingleDenom: []sdk.DecCoins{},
		hundredTokensMultiDenom: []sdk.DecCoins{},
		twoHundredTokensMultiDenom: []sdk.DecCoins{},
		varyingTokensSingleDenom: []sdk.DecCoins{},
		varyingTokensMultiDenom: []sdk.DecCoins{},
	}
	for i := range types.SupportedUptimes {
		expUptimes.emptyExpectedAccumValues = append(expUptimes.emptyExpectedAccumValues, cl.EmptyCoins)
		expUptimes.hundredTokensSingleDenom = append(expUptimes.hundredTokensSingleDenom, sdk.NewDecCoins(cl.HundredFooCoins))
		expUptimes.hundredTokensMultiDenom = append(expUptimes.hundredTokensMultiDenom, sdk.NewDecCoins(cl.HundredFooCoins, cl.HundredBarCoins))
		expUptimes.twoHundredTokensMultiDenom = append(expUptimes.twoHundredTokensMultiDenom, sdk.NewDecCoins(cl.HundredFooCoins.Add(cl.HundredFooCoins), cl.HundredBarCoins.Add(cl.HundredBarCoins)))
		expUptimes.varyingTokensSingleDenom = append(expUptimes.varyingTokensSingleDenom, sdk.NewDecCoins(cl.HundredFooCoins.Add(sdk.NewDecCoin("foo", sdk.NewInt(int64(i))))))
		expUptimes.varyingTokensMultiDenom = append(expUptimes.varyingTokensMultiDenom, sdk.NewDecCoins(cl.HundredFooCoins.Add(sdk.NewDecCoin("foo", sdk.NewInt(int64(i)))), cl.HundredBarCoins.Add(sdk.NewDecCoin("bar", sdk.NewInt(int64(i * 3))))))
	}

	return expUptimes
}

// Helper for converting raw DecCoins accum values to pool proto compatible UptimeTrackers
func wrapUptimeTrackers(accumValues []sdk.DecCoins) []model.UptimeTracker {
	wrappedUptimeTrackers := []model.UptimeTracker{}
	for _, accumValue := range accumValues {
		wrappedUptimeTrackers = append(wrappedUptimeTrackers, model.UptimeTracker{accumValue})
	}

	return wrappedUptimeTrackers
}

func expectedIncentives(denom string, rate sdk.Dec, timeElapsed time.Duration, qualifyingLiquidity sdk.Dec) sdk.DecCoin {
	amount := rate.Mul(sdk.NewDec(int64(timeElapsed))).Quo(qualifyingLiquidity)

	return sdk.NewDecCoinFromDec(denom, amount)
}

func chargeIncentive(incentiveRecord types.IncentiveRecord, timeElapsed time.Duration) types.IncentiveRecord {
	incentivesEmitted := incentiveRecord.EmissionRate.Mul(sdk.NewDec(int64(timeElapsed)))
	incentiveRecord.RemainingAmount = incentiveRecord.RemainingAmount.Sub(incentivesEmitted)

	return incentiveRecord
}

func (s *KeeperTestSuite) TestCreateAndGetUptimeAccumulators() {
	// We expect there to be len(types.SupportedUptimes) number of initialized accumulators
	// for a successful pool creation. We calculate this upfront to ensure test compatibility
	// if the uptimes we support ever change.
	expectedUptimes := getExpectedUptimes()

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
			expectedAccumValues: expectedUptimes.emptyExpectedAccumValues,
			expectedPass:        true,
		},
		"setup with different poolId": {
			poolId:              defaultPoolId + 1,
			initializePoolAccum: true,
			expectedAccumValues: expectedUptimes.emptyExpectedAccumValues,
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
		uptimeIndex       uint64
		expectedAccumName string
	}
	tests := map[string]getUptimeNameTest{
		"pool id 1, uptime id 0": {
			poolId:            defaultPoolId,
			uptimeIndex:       uint64(0),
			expectedAccumName: "uptime/1/0",
		},
		"pool id 1, uptime id 999": {
			poolId:            defaultPoolId,
			uptimeIndex:       uint64(999),
			expectedAccumName: "uptime/1/999",
		},
		"pool id 999, uptime id 1": {
			poolId:            uint64(999),
			uptimeIndex:       uint64(1),
			expectedAccumName: "uptime/999/1",
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()

			// system under test
			accumName := cl.GetUptimeAccumulatorName(tc.poolId, tc.uptimeIndex)
			s.Require().Equal(tc.expectedAccumName, accumName)
		})
	}
}

func (s *KeeperTestSuite) TestCreateAndGetUptimeAccumulatorValues() {
	// We expect there to be len(types.SupportedUptimes) number of initialized accumulators
	// for a successful pool creation. 
	// We re-calculate these values each time to ensure test compatibility if the uptimes 
	// we support ever change.
	expectedUptimes := getExpectedUptimes()

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
			addedAccumValues:	expectedUptimes.hundredTokensSingleDenom,
			numTimesAdded: 1,
			expectedAccumValues: expectedUptimes.hundredTokensSingleDenom,
			expectedPass:        true,
		},
		"hundred of multiple denom in each accumulator added once": {
			poolId:              defaultPoolId,
			initializePoolAccums:      true,
			addedAccumValues:	expectedUptimes.hundredTokensMultiDenom,
			numTimesAdded: 1,
			expectedAccumValues: expectedUptimes.hundredTokensMultiDenom,
			expectedPass:        true,
		},
		"varying amounts of single denom in each accumulator added once": {
			poolId:              defaultPoolId,
			initializePoolAccums:      true,
			addedAccumValues:	expectedUptimes.varyingTokensSingleDenom,
			numTimesAdded: 1,
			expectedAccumValues: expectedUptimes.varyingTokensSingleDenom,
			expectedPass:        true,
		},
		"varying of multiple denoms in each accumulator added once": {
			poolId:              defaultPoolId,
			initializePoolAccums:      true,
			addedAccumValues:	expectedUptimes.varyingTokensMultiDenom,
			numTimesAdded: 1,
			expectedAccumValues: expectedUptimes.varyingTokensMultiDenom,
			expectedPass:        true,
		},
		"hundred of multiple denom in each accumulator added twice": {
			poolId:              defaultPoolId,
			initializePoolAccums:      true,
			addedAccumValues:	expectedUptimes.hundredTokensMultiDenom,
			numTimesAdded: 2,
			expectedAccumValues: expectedUptimes.twoHundredTokensMultiDenom,
			expectedPass:        true,
		},
		"setup with different poolId": {
			poolId:              defaultPoolId + 1,
			initializePoolAccums:      true,
			addedAccumValues:	expectedUptimes.hundredTokensSingleDenom,
			numTimesAdded: 1,
			expectedAccumValues: expectedUptimes.hundredTokensSingleDenom,
			expectedPass:        true,
		},
		"pool not initialized": {
			initializePoolAccums:      false,
			poolId:              defaultPoolId,
			addedAccumValues:	expectedUptimes.hundredTokensSingleDenom,
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

func (s *KeeperTestSuite) TestCalcAccruedIncentivesForAccum() {
	type calcAccruedIncentivesTest struct {
		poolId              uint64
		accumUptime time.Duration
		qualifyingLiquidity sdk.Dec
		timeElapsed	time.Duration
		poolIncentiveRecords []types.IncentiveRecord

		expectedResult	sdk.DecCoins
		expectedIncentiveRecords []types.IncentiveRecord
		expectedPass bool
	}
	tests := map[string]calcAccruedIncentivesTest{
		"one incentive record, one qualifying for incentives": {
			poolId:              defaultPoolId,
			accumUptime: types.SupportedUptimes[0],
			qualifyingLiquidity: sdk.NewDec(100),
			timeElapsed: time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},

			expectedResult: sdk.DecCoins{
				expectedIncentives(incentiveRecordOne.IncentiveDenom, incentiveRecordOne.EmissionRate, time.Hour, sdk.NewDec(100)),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{chargeIncentive(incentiveRecordOne, time.Hour)},
			expectedPass:        true,
		},
		"two incentive records, one qualifying for incentives": {
			poolId:              defaultPoolId,
			accumUptime: types.SupportedUptimes[0],
			qualifyingLiquidity: sdk.NewDec(100),
			timeElapsed: time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo},

			expectedResult: sdk.DecCoins{
				// We only expect the first incentive record to qualify
				expectedIncentives(incentiveRecordOne.IncentiveDenom, incentiveRecordOne.EmissionRate, time.Hour, sdk.NewDec(100)),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{
				// We only charge the first incentive record since the second wasn't affected
				chargeIncentive(incentiveRecordOne, time.Hour),
				incentiveRecordTwo,
			},
			expectedPass:        true,
		},

		// error catching
		"zero qualifying liquidity": {
			poolId:              defaultPoolId,
			accumUptime: types.SupportedUptimes[0],
			qualifyingLiquidity: sdk.NewDec(0),
			timeElapsed: time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},

			expectedResult: sdk.DecCoins{},
			expectedIncentiveRecords: []types.IncentiveRecord{},
			expectedPass:        false,
		},
		"zero time elapsed": {
			poolId:              defaultPoolId,
			accumUptime: types.SupportedUptimes[0],
			qualifyingLiquidity: sdk.NewDec(100),
			timeElapsed: time.Duration(0),
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},

			expectedResult: sdk.DecCoins{},
			expectedIncentiveRecords: []types.IncentiveRecord{},
			expectedPass:        false,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultStartTime.Add(tc.timeElapsed))

			s.PrepareConcentratedPool()

			// system under test
			actualResult, updatedPoolRecords, err := cl.CalcAccruedIncentivesForAccum(s.Ctx, tc.accumUptime, tc.qualifyingLiquidity, sdk.NewDec(int64(tc.timeElapsed)), tc.poolIncentiveRecords)

			if tc.expectedPass {
				s.Require().NoError(err)

				s.Require().Equal(tc.expectedResult, actualResult)
				s.Require().Equal(tc.expectedIncentiveRecords, updatedPoolRecords)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
