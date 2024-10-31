package concentrated_liquidity_test

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
)

var (
	defaultPoolId     = uint64(1)
	defaultMultiplier = osmomath.OneInt()

	testAddressOne   = sdk.MustAccAddressFromBech32("osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks")
	testAddressTwo   = sdk.MustAccAddressFromBech32("osmo18s5lynnmx37hq4wlrw9gdn68sg2uxp5rgk26vv")
	testAddressThree = sdk.MustAccAddressFromBech32("osmo1qwexv7c6sm95lwhzn9027vyu2ccneaqad4w8ka")
	testAddressFour  = sdk.MustAccAddressFromBech32("osmo14hcxlnwlqtq75ttaxf674vk6mafspg8xwgnn53")

	// Note: lexicographic order is denomFour, denomOne, denomThree, denomTwo
	testDenomOne   = "denomOne"
	testDenomTwo   = "denomTwo"
	testDenomThree = "denomThree"
	testDenomFour  = "denomFour"

	defaultIncentiveAmount   = osmomath.NewDec(2 << 60)
	defaultIncentiveRecordId = uint64(1)

	testEmissionOne   = osmomath.MustNewDecFromStr("0.000001")
	testEmissionTwo   = osmomath.MustNewDecFromStr("0.0783")
	testEmissionThree = osmomath.MustNewDecFromStr("165.4")
	testEmissionFour  = osmomath.MustNewDecFromStr("57.93")

	defaultBlockTime  = time.Unix(1, 1).UTC()
	defaultTimeBuffer = time.Hour
	defaultStartTime  = defaultBlockTime.Add(defaultTimeBuffer)

	testUptimeOne   = types.SupportedUptimes[0]
	testUptimeTwo   = types.SupportedUptimes[1]
	testUptimeThree = types.SupportedUptimes[2]
	testUptimeFour  = types.SupportedUptimes[3]

	incentiveRecordOne = types.IncentiveRecord{
		PoolId: validPoolId,
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingCoin: sdk.NewDecCoinFromDec(testDenomOne, defaultIncentiveAmount),
			EmissionRate:  testEmissionOne,
			StartTime:     defaultStartTime,
		},
		MinUptime:   testUptimeOne,
		IncentiveId: defaultIncentiveRecordId,
	}

	incentiveRecordTwo = types.IncentiveRecord{
		PoolId: validPoolId,
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingCoin: sdk.NewDecCoinFromDec(testDenomTwo, defaultIncentiveAmount),
			EmissionRate:  testEmissionTwo,
			StartTime:     defaultStartTime,
		},
		MinUptime:   testUptimeTwo,
		IncentiveId: defaultIncentiveRecordId + 1,
	}

	incentiveRecordThree = types.IncentiveRecord{
		PoolId: validPoolId,
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingCoin: sdk.NewDecCoinFromDec(testDenomThree, defaultIncentiveAmount),
			EmissionRate:  testEmissionThree,
			StartTime:     defaultStartTime,
		},
		MinUptime:   testUptimeThree,
		IncentiveId: defaultIncentiveRecordId + 2,
	}

	incentiveRecordFour = types.IncentiveRecord{
		PoolId: validPoolId,
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingCoin: sdk.NewDecCoinFromDec(testDenomFour, defaultIncentiveAmount),
			EmissionRate:  testEmissionFour,
			StartTime:     defaultStartTime,
		},
		MinUptime:   testUptimeFour,
		IncentiveId: defaultIncentiveRecordId + 3,
	}

	emptyIncentiveRecord = types.IncentiveRecord{
		PoolId: validPoolId,
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingCoin: sdk.NewDecCoinFromDec("emptyDenom", osmomath.ZeroDec()),
			EmissionRate:  testEmissionFour,
			StartTime:     defaultStartTime,
		},
		MinUptime:   testUptimeFour,
		IncentiveId: defaultIncentiveRecordId + 4,
	}

	testQualifyingDepositsOne = osmomath.NewInt(50)

	defaultBalancerAssets = []balancer.PoolAsset{
		{Weight: osmomath.NewInt(1), Token: sdk.NewCoin("foo", osmomath.NewInt(100))},
		{Weight: osmomath.NewInt(1), Token: sdk.NewCoin("bar", osmomath.NewInt(100))},
	}
	defaultConcentratedAssets = sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(100)), sdk.NewCoin("bar", osmomath.NewInt(100)))
	defaultBalancerPoolParams = balancer.PoolParams{SwapFee: osmomath.NewDec(0), ExitFee: osmomath.NewDec(0)}
	invalidPoolId             = uint64(10)

	// 10^60
	oneE60Dec = osmomath.MustNewDecFromStr("1000000000000000000000000000000000000000000000000000000000000")
)

type ExpectedUptimes struct {
	emptyExpectedAccumValues     []sdk.DecCoins
	fiveHundredAccumValues       []sdk.DecCoins
	hundredTokensSingleDenom     []sdk.DecCoins
	hundredTokensMultiDenom      []sdk.DecCoins
	twoHundredTokensMultiDenom   []sdk.DecCoins
	threeHundredTokensMultiDenom []sdk.DecCoins
	fourHundredTokensMultiDenom  []sdk.DecCoins
	sixHundredTokensMultiDenom   []sdk.DecCoins
	varyingTokensSingleDenom     []sdk.DecCoins
	varyingTokensMultiDenom      []sdk.DecCoins
}

// getExpectedUptimes returns a base set of expected values for testing based on the number
// of supported uptimes at runtime. This abstraction exists only to ensure backwards-compatibility
// of incentives-related tests if the supported uptimes are ever changed.
func getExpectedUptimes() ExpectedUptimes {
	expUptimes := ExpectedUptimes{
		emptyExpectedAccumValues:     []sdk.DecCoins{},
		hundredTokensSingleDenom:     []sdk.DecCoins{},
		hundredTokensMultiDenom:      []sdk.DecCoins{},
		twoHundredTokensMultiDenom:   []sdk.DecCoins{},
		threeHundredTokensMultiDenom: []sdk.DecCoins{},
		fourHundredTokensMultiDenom:  []sdk.DecCoins{},
		sixHundredTokensMultiDenom:   []sdk.DecCoins{},
		varyingTokensSingleDenom:     []sdk.DecCoins{},
		varyingTokensMultiDenom:      []sdk.DecCoins{},
	}
	for i := range types.SupportedUptimes {
		expUptimes.emptyExpectedAccumValues = append(expUptimes.emptyExpectedAccumValues, cl.EmptyCoins)
		expUptimes.hundredTokensSingleDenom = append(expUptimes.hundredTokensSingleDenom, sdk.NewDecCoins(cl.HundredFooCoins))
		expUptimes.hundredTokensMultiDenom = append(expUptimes.hundredTokensMultiDenom, sdk.NewDecCoins(cl.HundredFooCoins, cl.HundredBarCoins))
		expUptimes.twoHundredTokensMultiDenom = append(expUptimes.twoHundredTokensMultiDenom, sdk.NewDecCoins(cl.TwoHundredFooCoins, cl.TwoHundredBarCoins))
		expUptimes.threeHundredTokensMultiDenom = append(expUptimes.threeHundredTokensMultiDenom, sdk.NewDecCoins(cl.TwoHundredFooCoins.Add(cl.HundredFooCoins), cl.TwoHundredBarCoins.Add(cl.HundredBarCoins)))
		expUptimes.fourHundredTokensMultiDenom = append(expUptimes.fourHundredTokensMultiDenom, sdk.NewDecCoins(cl.TwoHundredFooCoins.Add(cl.TwoHundredFooCoins), cl.TwoHundredBarCoins.Add(cl.TwoHundredBarCoins)))
		expUptimes.sixHundredTokensMultiDenom = append(expUptimes.sixHundredTokensMultiDenom, sdk.NewDecCoins(cl.TwoHundredFooCoins.Add(cl.TwoHundredFooCoins).Add(cl.TwoHundredFooCoins), cl.TwoHundredBarCoins.Add(cl.TwoHundredBarCoins).Add(cl.TwoHundredBarCoins)))
		expUptimes.varyingTokensSingleDenom = append(expUptimes.varyingTokensSingleDenom, sdk.NewDecCoins(cl.HundredFooCoins.Add(sdk.NewDecCoin("foo", osmomath.NewInt(int64(i))))))
		expUptimes.varyingTokensMultiDenom = append(expUptimes.varyingTokensMultiDenom, sdk.NewDecCoins(cl.HundredFooCoins.Add(sdk.NewDecCoin("foo", osmomath.NewInt(int64(i)))), cl.HundredBarCoins.Add(sdk.NewDecCoin("bar", osmomath.NewInt(int64(i*3))))))
	}

	return expUptimes
}

// Helper for converting raw DecCoins accum values to pool proto compatible UptimeTrackers
func wrapUptimeTrackers(accumValues []sdk.DecCoins) []model.UptimeTracker {
	wrappedUptimeTrackers := []model.UptimeTracker{}
	for _, accumValue := range accumValues {
		wrappedUptimeTrackers = append(wrappedUptimeTrackers, model.UptimeTracker{UptimeGrowthOutside: accumValue})
	}

	return wrappedUptimeTrackers
}

// expectedIncentivesFromRate calculates the amount of incentives we expect to accrue based on the rate and time elapsed
func expectedIncentivesFromRate(denom string, rate osmomath.Dec, timeElapsed time.Duration, qualifyingLiquidity osmomath.Dec) sdk.DecCoin {
	timeInSec := osmomath.NewDec(int64(timeElapsed)).Quo(osmomath.MustNewDecFromStr("1000000000")).MulTruncateMut(cl.PerUnitLiqScalingFactor)
	amount := rate.Mul(timeInSec).QuoTruncate(qualifyingLiquidity)

	return sdk.NewDecCoinFromDec(denom, amount)
}

// expectedIncentivesFromUptimeGrowth calculates the amount of incentives we expect to accrue based on uptime accumulator growth.
//
// Assumes `uptimeGrowths` represents the growths for all global uptime accums and only counts growth that `timeInPool` qualifies for
// towards result. Takes in a multiplier parameter for further versatility in testing.
//
// Returns value as truncated sdk.Coins as the primary use of this helper is testing higher level incentives functions such as claiming.
func expectedIncentivesFromUptimeGrowth(uptimeGrowths []sdk.DecCoins, positionShares osmomath.Dec, timeInPool time.Duration, multiplier osmomath.Int) sdk.Coins {
	// Sum up rewards from all inputs
	totalRewards := sdk.Coins(nil)
	for uptimeIndex, uptimeGrowth := range uptimeGrowths {
		if timeInPool >= types.SupportedUptimes[uptimeIndex] {
			curRewards := uptimeGrowth.MulDecTruncate(positionShares).MulDecTruncate(multiplier.ToLegacyDec())
			totalRewards = totalRewards.Add(sdk.NormalizeCoins(curRewards)...)
		}
	}

	return totalRewards
}

// chargeIncentiveRecord updates the remaining amount of the passed in incentive record to what it would be after `timeElapsed` of emissions.
func chargeIncentiveRecord(incentiveRecord types.IncentiveRecord, timeElapsed time.Duration) types.IncentiveRecord {
	secToNanoSec := int64(1000000000)
	incentivesEmitted := incentiveRecord.IncentiveRecordBody.EmissionRate.Mul(osmomath.NewDec(int64(timeElapsed)).Quo(osmomath.NewDec(secToNanoSec)))
	incentiveRecord.IncentiveRecordBody.RemainingCoin.Amount = incentiveRecord.IncentiveRecordBody.RemainingCoin.Amount.Sub(incentivesEmitted)

	return incentiveRecord
}

// Helper for adding a predetermined amount to each global uptime accum in clPool
func addToUptimeAccums(ctx sdk.Context, poolId uint64, clKeeper *cl.Keeper, addValues []sdk.DecCoins) error {
	poolUptimeAccumulators, err := clKeeper.GetUptimeAccumulators(ctx, poolId)
	if err != nil {
		return err
	}

	for uptimeIndex, uptimeAccum := range poolUptimeAccumulators {
		uptimeAccum.AddToAccumulator(addValues[uptimeIndex].MulDecTruncate(cl.PerUnitLiqScalingFactor))
	}

	return nil
}

func withDenom(record types.IncentiveRecord, denom string) types.IncentiveRecord {
	record.IncentiveRecordBody.RemainingCoin.Denom = denom

	return record
}

func withAmount(record types.IncentiveRecord, amount osmomath.Dec) types.IncentiveRecord {
	record.IncentiveRecordBody.RemainingCoin.Amount = amount

	return record
}

func withStartTime(record types.IncentiveRecord, startTime time.Time) types.IncentiveRecord {
	record.IncentiveRecordBody.StartTime = startTime

	return record
}

func withMinUptime(record types.IncentiveRecord, minUptime time.Duration) types.IncentiveRecord {
	record.MinUptime = minUptime

	return record
}

func withEmissionRate(record types.IncentiveRecord, emissionRate osmomath.Dec) types.IncentiveRecord {
	record.IncentiveRecordBody.EmissionRate = emissionRate

	return record
}

// TestCreateAndGetUptimeAccumulators tests the creation and retrieval logic for pool-wide uptime accumulators.
// Note that this is distinct from the authorized uptime accumulators, which will always be a subset of the accumulators
// considered in this test.
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
	s.runMultipleAuthorizedUptimes(func() {
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

					// Ensure number of uptime accumulators match supported uptimes
					s.Require().Equal(len(tc.expectedAccumValues), len(poolUptimeAccumulators))

					// Ensure that each uptime was initialized to the correct value (sdk.DecCoins(nil))
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
	})
}

// TestGetUptimeAccumulatorName tests the name generation logic for pool-wide uptime accumulators.
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
			expectedAccumName: types.KeyUptimeAccumulator(1, 0),
		},
		"pool id 1, uptime id 999": {
			poolId:            defaultPoolId,
			uptimeIndex:       uint64(999),
			expectedAccumName: types.KeyUptimeAccumulator(1, 999),
		},
		"pool id 999, uptime id 1": {
			poolId:            uint64(999),
			uptimeIndex:       uint64(1),
			expectedAccumName: types.KeyUptimeAccumulator(999, 1),
		},
	}

	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			tc := tc
			s.Run(name, func() {
				s.SetupTest()

				// system under test
				accumName := types.KeyUptimeAccumulator(tc.poolId, tc.uptimeIndex)
				s.Require().Equal(tc.expectedAccumName, accumName)
			})
		}
	})
}

// TestCreateAndGetUptimeAccumulatorValues tests the creation and retrieval logic for pool-wide uptime accumulator values.
// Note that this is independent of authorized uptimes as the functions in question deal with all the active accumulators on the pool.
func (s *KeeperTestSuite) TestCreateAndGetUptimeAccumulatorValues() {
	// We expect there to be len(types.SupportedUptimes) number of initialized accumulators
	// for a successful pool creation.
	//
	// We re-calculate these values each time to ensure test compatibility if the uptimes
	// we support ever change.
	expectedUptimes := getExpectedUptimes()

	type initUptimeAccumTest struct {
		poolId               uint64
		initializePoolAccums bool
		addedAccumValues     []sdk.DecCoins
		numTimesAdded        int
		expectedAccumValues  []sdk.DecCoins

		expectedPass bool
	}
	tests := map[string]initUptimeAccumTest{
		"hundred of a single denom in each accumulator added once": {
			poolId:               defaultPoolId,
			initializePoolAccums: true,
			addedAccumValues:     expectedUptimes.hundredTokensSingleDenom,
			numTimesAdded:        1,
			expectedAccumValues:  expectedUptimes.hundredTokensSingleDenom,
			expectedPass:         true,
		},
		"hundred of multiple denom in each accumulator added once": {
			poolId:               defaultPoolId,
			initializePoolAccums: true,
			addedAccumValues:     expectedUptimes.hundredTokensMultiDenom,
			numTimesAdded:        1,
			expectedAccumValues:  expectedUptimes.hundredTokensMultiDenom,
			expectedPass:         true,
		},
		"varying amounts of single denom in each accumulator added once": {
			poolId:               defaultPoolId,
			initializePoolAccums: true,
			addedAccumValues:     expectedUptimes.varyingTokensSingleDenom,
			numTimesAdded:        1,
			expectedAccumValues:  expectedUptimes.varyingTokensSingleDenom,
			expectedPass:         true,
		},
		"varying of multiple denoms in each accumulator added once": {
			poolId:               defaultPoolId,
			initializePoolAccums: true,
			addedAccumValues:     expectedUptimes.varyingTokensMultiDenom,
			numTimesAdded:        1,
			expectedAccumValues:  expectedUptimes.varyingTokensMultiDenom,
			expectedPass:         true,
		},
		"hundred of multiple denom in each accumulator added twice": {
			poolId:               defaultPoolId,
			initializePoolAccums: true,
			addedAccumValues:     expectedUptimes.hundredTokensMultiDenom,
			numTimesAdded:        2,
			expectedAccumValues:  expectedUptimes.twoHundredTokensMultiDenom,
			expectedPass:         true,
		},
		"setup with different poolId": {
			poolId:               defaultPoolId + 1,
			initializePoolAccums: true,
			addedAccumValues:     expectedUptimes.hundredTokensSingleDenom,
			numTimesAdded:        1,
			expectedAccumValues:  expectedUptimes.hundredTokensSingleDenom,
			expectedPass:         true,
		},
		"pool not initialized": {
			initializePoolAccums: false,
			poolId:               defaultPoolId,
			addedAccumValues:     expectedUptimes.hundredTokensSingleDenom,
			numTimesAdded:        1,
			expectedAccumValues:  []sdk.DecCoins{},
			expectedPass:         false,
		},
	}

	s.runMultipleAuthorizedUptimes(func() {
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
	})
}

func (s *KeeperTestSuite) getAllIncentiveRecordsForPool(poolId uint64) []types.IncentiveRecord {
	records, err := s.Clk.GetAllIncentiveRecordsForPool(s.Ctx, poolId)
	s.Require().NoError(err)
	return records
}

// TestCalcAccruedIncentivesForAccum tests the calculation of accrued incentives for a specific accumulator on a pool.
func (s *KeeperTestSuite) TestCalcAccruedIncentivesForAccum() {
	incentiveRecordOneWithDifferentStartTime := withStartTime(incentiveRecordOne, incentiveRecordOne.IncentiveRecordBody.StartTime.Add(10))
	incentiveRecordOneWithDifferentMinUpTime := withMinUptime(incentiveRecordOne, testUptimeTwo)
	incentiveRecordOneWithDifferentDenom := withDenom(incentiveRecordOne, testDenomTwo)
	incentiveRecordOneWithStartTimeAfterBlockTime := withStartTime(incentiveRecordOne, incentiveRecordOne.IncentiveRecordBody.StartTime.Add(time.Hour*24))

	type calcAccruedIncentivesTest struct {
		poolId               uint64
		accumUptime          time.Duration
		qualifyingLiquidity  osmomath.Dec
		timeElapsed          time.Duration
		poolIncentiveRecords []types.IncentiveRecord
		recordsCleared       bool

		expectedResult           sdk.DecCoins
		expectedIncentiveRecords []types.IncentiveRecord
		expectedPass             bool
	}
	tests := map[string]calcAccruedIncentivesTest{
		"one incentive record, one qualifying for incentives": {
			poolId:               defaultPoolId,
			accumUptime:          types.SupportedUptimes[0],
			qualifyingLiquidity:  hundredDec,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},

			expectedResult: sdk.DecCoins{
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, hundredDec),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{chargeIncentiveRecord(incentiveRecordOne, time.Hour)},
			expectedPass:             true,
		},
		"one incentive record, one qualifying for incentives, start time after current block time": {
			poolId:               defaultPoolId,
			accumUptime:          types.SupportedUptimes[0],
			qualifyingLiquidity:  hundredDec,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOneWithStartTimeAfterBlockTime},

			expectedResult:           sdk.DecCoins{},
			expectedIncentiveRecords: []types.IncentiveRecord{incentiveRecordOneWithStartTimeAfterBlockTime},
			expectedPass:             true,
		},
		"two incentive records, one with qualifying liquidity for incentives": {
			poolId:               defaultPoolId,
			accumUptime:          types.SupportedUptimes[0],
			qualifyingLiquidity:  hundredDec,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo},

			expectedResult: sdk.DecCoins{
				// We only expect the first incentive record to qualify
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, hundredDec),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{
				// We only charge the first incentive record since the second wasn't affected
				chargeIncentiveRecord(incentiveRecordOne, time.Hour),
				incentiveRecordTwo,
			},
			expectedPass: true,
		},
		"fully emit all incentives in record, significant time elapsed": {
			poolId:              defaultPoolId,
			accumUptime:         types.SupportedUptimes[0],
			qualifyingLiquidity: osmomath.NewDec(123),

			// Time elapsed is strictly greater than the time needed to emit all incentives
			timeElapsed: time.Duration((1 << 63) - 1),
			poolIncentiveRecords: []types.IncentiveRecord{
				// We set the emission rate high enough to drain the record in one timestep
				withEmissionRate(incentiveRecordOne, osmomath.NewDec(2<<60)),
			},
			recordsCleared: true,

			// We expect the fully incentive amount to be emitted
			expectedResult: sdk.DecCoins{
				sdk.NewDecCoinFromDec(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.MulTruncate(cl.PerUnitLiqScalingFactor).QuoTruncate(osmomath.NewDec(123))),
			},

			// Incentive record should have zero remaining amountosmomath.ZeroDec
			expectedIncentiveRecords: []types.IncentiveRecord{withAmount(withEmissionRate(incentiveRecordOne, osmomath.NewDec(2<<60)), osmomath.ZeroDec())},
			expectedPass:             true,
		},

		"two incentive records, first overflows, second still succeeds": {
			poolId:               defaultPoolId,
			accumUptime:          types.SupportedUptimes[0],
			qualifyingLiquidity:  hundredDec,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{withEmissionRate(incentiveRecordOne, oneE60Dec), incentiveRecordOne},

			expectedResult: sdk.DecCoins{
				// We expect both incentives to qualify. However, only the second one emits because
				// the first one is truncated.
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, hundredDec),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{
				withEmissionRate(incentiveRecordOne, oneE60Dec),
				// We only charge the second incentive record since the first silently errored due to overflow.
				chargeIncentiveRecord(incentiveRecordOne, time.Hour),
			},
			expectedPass: true,
		},

		// error catching
		"zero qualifying liquidity": {
			poolId:               defaultPoolId,
			accumUptime:          types.SupportedUptimes[0],
			qualifyingLiquidity:  osmomath.NewDec(0),
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},

			expectedResult:           sdk.DecCoins{},
			expectedIncentiveRecords: []types.IncentiveRecord{},
			expectedPass:             false,
		},
		"zero time elapsed": {
			poolId:               defaultPoolId,
			accumUptime:          types.SupportedUptimes[0],
			qualifyingLiquidity:  hundredDec,
			timeElapsed:          time.Duration(0),
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},

			expectedResult:           sdk.DecCoins{},
			expectedIncentiveRecords: []types.IncentiveRecord{},
			expectedPass:             false,
		},
		"two incentive records with same denom, different start time": {
			poolId:              defaultPoolId,
			accumUptime:         types.SupportedUptimes[0],
			qualifyingLiquidity: hundredDec,
			timeElapsed:         time.Hour,

			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordOneWithDifferentStartTime},

			expectedResult: sdk.NewDecCoins(
				// We expect both incentive records to qualify
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate.Add(incentiveRecordOneWithDifferentStartTime.IncentiveRecordBody.EmissionRate), time.Hour, hundredDec), // since we have 2 records with same denom, the rate of emission went up x2
			),
			expectedIncentiveRecords: []types.IncentiveRecord{
				// We only going to charge both incentive records
				chargeIncentiveRecord(incentiveRecordOne, time.Hour),
				chargeIncentiveRecord(incentiveRecordOneWithDifferentStartTime, time.Hour),
			},
			expectedPass: true,
		},
		"two incentive records with different denom, different start time and same uptime": {
			poolId:              defaultPoolId,
			accumUptime:         types.SupportedUptimes[0],
			qualifyingLiquidity: hundredDec,
			timeElapsed:         time.Hour,

			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOneWithDifferentStartTime, incentiveRecordOneWithDifferentDenom},

			expectedResult: sdk.DecCoins{
				// We expect both incentive record to qualify
				expectedIncentivesFromRate(incentiveRecordOneWithDifferentStartTime.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, hundredDec),
				expectedIncentivesFromRate(incentiveRecordOneWithDifferentDenom.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, hundredDec),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{
				// We charge both incentive record here because both minUpTime has been hit
				chargeIncentiveRecord(incentiveRecordOneWithDifferentStartTime, time.Hour),
				chargeIncentiveRecord(incentiveRecordOneWithDifferentDenom, time.Hour),
			},
			expectedPass: true,
		},
		"two incentive records with same denom, different min up-time": {
			poolId:              defaultPoolId,
			accumUptime:         types.SupportedUptimes[0],
			qualifyingLiquidity: hundredDec,
			timeElapsed:         time.Hour,

			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordOneWithDifferentMinUpTime},

			expectedResult: sdk.DecCoins{
				// We expect first incentive record to qualify
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, hundredDec),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{
				// We only charge the first incentive record because the second minUpTime hasn't been hit yet
				chargeIncentiveRecord(incentiveRecordOne, time.Hour),
				incentiveRecordOneWithDifferentMinUpTime,
			},
			expectedPass: true,
		},
		"two incentive records with same accum uptime and start time across multiple records with different denoms": {
			poolId:              defaultPoolId,
			accumUptime:         types.SupportedUptimes[0],
			qualifyingLiquidity: hundredDec,
			timeElapsed:         time.Hour,

			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordOneWithDifferentDenom},

			expectedResult: sdk.DecCoins{
				// We expect both incentive record to qualify
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, hundredDec),
				expectedIncentivesFromRate(incentiveRecordOneWithDifferentDenom.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, hundredDec),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{
				// We charge both incentive record here because both minUpTime has been hit
				chargeIncentiveRecord(incentiveRecordOne, time.Hour),
				chargeIncentiveRecord(incentiveRecordOneWithDifferentDenom, time.Hour),
			},
			expectedPass: true,
		},
		"four incentive records with only two eligilbe for emitting incentives": {
			poolId:              defaultPoolId,
			accumUptime:         types.SupportedUptimes[0],
			qualifyingLiquidity: hundredDec,
			timeElapsed:         time.Hour,

			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordOneWithDifferentStartTime, incentiveRecordOneWithDifferentDenom, incentiveRecordOneWithDifferentMinUpTime},

			expectedResult: sdk.NewDecCoins(
				// We expect three incentive record to qualify for incentive
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate.Add(incentiveRecordOneWithDifferentStartTime.IncentiveRecordBody.EmissionRate), time.Hour, hundredDec),
				expectedIncentivesFromRate(incentiveRecordOneWithDifferentDenom.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, hundredDec),
			),
			expectedIncentiveRecords: []types.IncentiveRecord{
				// We only charge the first three incentive record because the fourth minUpTime hasn't been hit yet
				chargeIncentiveRecord(incentiveRecordOne, time.Hour),
				chargeIncentiveRecord(incentiveRecordOneWithDifferentStartTime, time.Hour),
				chargeIncentiveRecord(incentiveRecordOneWithDifferentDenom, time.Hour),
				incentiveRecordOneWithDifferentMinUpTime, // this uptime hasn't hit yet so we do not have to charge incentive
			},
			expectedPass: true,
		},
	}

	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			tc := tc
			s.Run(name, func() {
				s.SetupTest()
				s.Ctx = s.Ctx.WithBlockTime(defaultStartTime.Add(tc.timeElapsed))

				s.PrepareConcentratedPool()

				// system under test
				actualResult, updatedPoolRecords, err := cl.CalcAccruedIncentivesForAccum(s.Ctx, tc.accumUptime, tc.qualifyingLiquidity, osmomath.NewDec(int64(tc.timeElapsed)).Quo(osmomath.MustNewDecFromStr("1000000000")), tc.poolIncentiveRecords, cl.PerUnitLiqScalingFactor)
				if tc.expectedPass {
					s.Require().NoError(err)

					s.Require().Equal(tc.expectedResult, actualResult)
					s.Require().Equal(tc.expectedIncentiveRecords, updatedPoolRecords)

					// If incentives are fully emitted, we ensure they are cleared from state
					if tc.recordsCleared {
						err := s.Clk.SetMultipleIncentiveRecords(s.Ctx, updatedPoolRecords)
						s.Require().NoError(err)

						updatedRecordsInState := s.getAllIncentiveRecordsForPool(tc.poolId)
						s.Require().Equal(0, len(updatedRecordsInState))
					}
				} else {
					s.Require().Error(err)
				}
			})
		}
	})
}

func (s *KeeperTestSuite) setupBalancerPoolWithFractionLocked(pa []balancer.PoolAsset, fraction osmomath.Dec) uint64 {
	balancerPoolId := s.PrepareCustomBalancerPool(pa, defaultBalancerPoolParams)
	longestLockableDuration, err := s.App.PoolIncentivesKeeper.GetLongestLockableDuration(s.Ctx)
	s.Require().NoError(err)
	lockAmt := gammtypes.InitPoolSharesSupply.ToLegacyDec().Mul(fraction).TruncateInt()
	lockCoins := sdk.NewCoins(sdk.NewCoin(gammtypes.GetPoolShareDenom(balancerPoolId), lockAmt))
	_, err = s.App.LockupKeeper.CreateLock(s.Ctx, s.TestAccs[0], lockCoins, longestLockableDuration)
	s.Require().NoError(err)
	return balancerPoolId
}

// Testing strategy:
// 1. Create a position
// 2. Let a fixed amount of time pass, enough to qualify it for some (but not all) uptimes
// 3. Let a variable amount of time pass determined by the test case
// 4. Ensure uptime accumulators and incentive records behave as expected
func (s *KeeperTestSuite) TestUpdateUptimeAccumulatorsToNow() {
	defaultTestUptime := types.SupportedUptimes[2]
	type updateAccumToNow struct {
		poolId               uint64
		timeElapsed          time.Duration
		poolIncentiveRecords []types.IncentiveRecord

		expectedIncentiveRecords []types.IncentiveRecord
		expectedError            error
	}

	validateResult := func(ctx sdk.Context, err error, tc updateAccumToNow, poolId uint64, initUptimeAccumValues []sdk.DecCoins, qualifyingLiquidity osmomath.Dec) []sdk.DecCoins {
		if tc.expectedError != nil {
			s.Require().ErrorContains(err, tc.expectedError.Error())

			// Ensure accumulators remain unchanged
			newUptimeAccumValues, err := s.Clk.GetUptimeAccumulatorValues(ctx, poolId)
			s.Require().NoError(err)
			s.Require().Equal(initUptimeAccumValues, newUptimeAccumValues)

			// Ensure incentive records remain unchanged
			updatedIncentiveRecords := s.getAllIncentiveRecordsForPool(poolId)
			s.Require().Equal(tc.poolIncentiveRecords, updatedIncentiveRecords)

			return nil
		}

		s.Require().NoError(err)

		// Get updated pool for testing purposes
		clPool, err := s.Clk.GetPoolById(ctx, tc.poolId)
		s.Require().NoError(err)

		// Calculate expected uptime deltas using qualifying liquidity deltas.
		// Recall that uptime accumulators track emitted incentives / qualifying liquidity.
		// Note: we test on all supported uptimes to ensure robustness, since even if only a subset is authorized they all technically get updated.
		expectedUptimeDeltas := []sdk.DecCoins{}
		for _, curSupportedUptime := range types.SupportedUptimes {
			// Calculate expected incentives for the current uptime by emitting incentives from
			// all incentive records to their respective uptime accumulators in the pool.
			// TODO: find a cleaner way to calculate this that does not involve iterating over all incentive records for each uptime accum.
			curUptimeAccruedIncentives := cl.EmptyCoins
			for _, poolRecord := range tc.poolIncentiveRecords {
				if poolRecord.MinUptime == curSupportedUptime {
					// We set the expected accrued incentives based on the total time that has elapsed since position creation
					curUptimeAccruedIncentives = curUptimeAccruedIncentives.Add(sdk.NewDecCoins(expectedIncentivesFromRate(poolRecord.IncentiveRecordBody.RemainingCoin.Denom, poolRecord.IncentiveRecordBody.EmissionRate, defaultTestUptime+tc.timeElapsed, qualifyingLiquidity))...)
				}
			}
			expectedUptimeDeltas = append(expectedUptimeDeltas, curUptimeAccruedIncentives)
		}

		// Get new uptime accum values for comparison
		newUptimeAccumValues, err := s.Clk.GetUptimeAccumulatorValues(ctx, tc.poolId)
		s.Require().NoError(err)

		// Ensure that each accumulator value changes by the correct amount
		totalUptimeDeltas := sdk.NewDecCoins()
		for uptimeIndex := range newUptimeAccumValues {
			uptimeDelta := newUptimeAccumValues[uptimeIndex].Sub(initUptimeAccumValues[uptimeIndex])
			s.Require().Equal(expectedUptimeDeltas[uptimeIndex], uptimeDelta)

			totalUptimeDeltas = totalUptimeDeltas.Add(uptimeDelta...)
		}

		// Ensure that LastLiquidityUpdate field is updated for pool
		s.Require().Equal(ctx.BlockTime(), clPool.GetLastLiquidityUpdate())

		// Ensure that pool's IncentiveRecords are updated to reflect emitted incentives
		updatedIncentiveRecords, err := s.Clk.GetAllIncentiveRecordsForPool(ctx, tc.poolId)
		s.Require().NoError(err)
		s.Require().Equal(tc.expectedIncentiveRecords, updatedIncentiveRecords)

		return expectedUptimeDeltas
	}

	tests := map[string]updateAccumToNow{
		"one incentive record": {
			poolId:               defaultPoolId,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We deduct incentives from the record for the period it emitted incentives
				chargeIncentiveRecord(incentiveRecordOne, defaultTestUptime+time.Hour),
			},
		},
		"two incentive records, each with qualifying liquidity": {
			poolId:               defaultPoolId,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo},

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We deduct incentives from both records since there are positions for each
				chargeIncentiveRecord(incentiveRecordOne, defaultTestUptime+time.Hour),
				chargeIncentiveRecord(incentiveRecordTwo, defaultTestUptime+time.Hour),
			},
		},
		"three incentive records, each with qualifying liquidity": {
			poolId:               defaultPoolId,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo, incentiveRecordThree},

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We deduct incentives from each record since there are positions for all three
				// Note that records are in ascending order by uptime index
				chargeIncentiveRecord(incentiveRecordOne, defaultTestUptime+time.Hour),
				chargeIncentiveRecord(incentiveRecordTwo, defaultTestUptime+time.Hour),
				chargeIncentiveRecord(incentiveRecordThree, defaultTestUptime+time.Hour),
			},
		},
		"four incentive records, only three with qualifying liquidity": {
			poolId:               defaultPoolId,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo, incentiveRecordThree, incentiveRecordFour},

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We only deduct from the first three incentive records since the last doesn't emit anything
				// Note that records are in ascending order by uptime index
				chargeIncentiveRecord(incentiveRecordOne, defaultTestUptime+time.Hour),
				chargeIncentiveRecord(incentiveRecordTwo, defaultTestUptime+time.Hour),
				chargeIncentiveRecord(incentiveRecordThree, defaultTestUptime+time.Hour),
				// We charge even for uptimes the position has technically not qualified for since its liquidity is on
				// the accumulator.
				chargeIncentiveRecord(incentiveRecordFour, defaultTestUptime+time.Hour),
			},
		},

		// Error catching

		"invalid pool ID": {
			poolId:               invalidPoolId,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We deduct incentives from the record for the period it emitted incentives
				chargeIncentiveRecord(incentiveRecordOne, defaultTestUptime+time.Hour),
			},
			expectedError: types.PoolNotFoundError{PoolId: invalidPoolId},
		},
	}

	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			tc := tc
			s.Run(name, func() {
				s.SetupTest()
				clKeeper := s.App.ConcentratedLiquidityKeeper
				s.Ctx = s.Ctx.WithBlockTime(defaultStartTime)

				// Set up test pool
				clPool := s.PrepareConcentratedPool()

				// Initialize test incentives on the pool.
				err := clKeeper.SetMultipleIncentiveRecords(s.Ctx, tc.poolIncentiveRecords)
				s.Require().NoError(err)

				// Get initial uptime accum values for comparison
				initUptimeAccumValues, err := clKeeper.GetUptimeAccumulatorValues(s.Ctx, clPool.GetId())
				s.Require().NoError(err)

				// Add qualifying and non-qualifying liquidity to the pool
				qualifyingLiquidity := osmomath.ZeroDec()
				depositedCoins := sdk.NewCoins(sdk.NewCoin(clPool.GetToken0(), testQualifyingDepositsOne), sdk.NewCoin(clPool.GetToken1(), testQualifyingDepositsOne))
				s.FundAcc(testAddressOne, depositedCoins)
				positionData, err := clKeeper.CreatePosition(s.Ctx, clPool.GetId(), testAddressOne, depositedCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), clPool.GetCurrentTick()-100, clPool.GetCurrentTick()+100)
				s.Require().NoError(err)
				qualifyingLiquidity = positionData.Liquidity

				clPool, err = clKeeper.GetPoolById(s.Ctx, clPool.GetId())
				s.Require().NoError(err)

				// Let enough time elapse to qualify the position for the first three supported uptimes
				s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(defaultTestUptime))

				// Let `timeElapsed` time pass to test incentive distribution
				s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(tc.timeElapsed))

				// System under test 1
				// Use cache context to avoid persisting updates for the next function
				// that relies on the same test cases and setup.
				cacheCtx, _ := s.Ctx.CacheContext()
				err = clKeeper.UpdatePoolUptimeAccumulatorsToNow(cacheCtx, tc.poolId)

				validateResult(cacheCtx, err, tc, clPool.GetId(), initUptimeAccumValues, qualifyingLiquidity)

				// Now test a similar method with different parameters

				// Skip this test case as UpdatePoolGivenUptimeAccumulatorsToNow relies
				// on this check to be done by the caller.
				if errors.Is(tc.expectedError, types.PoolNotFoundError{PoolId: invalidPoolId}) {
					return
				}

				uptimeAccs, err := clKeeper.GetUptimeAccumulators(s.Ctx, tc.poolId)
				s.Require().NoError(err)

				// System under test 2
				err = clKeeper.UpdateGivenPoolUptimeAccumulatorsToNow(s.Ctx, clPool, uptimeAccs)

				expectedUptimeDeltas := validateResult(s.Ctx, err, tc, clPool.GetId(), initUptimeAccumValues, qualifyingLiquidity)

				if tc.expectedError != nil {
					return
				}

				// Ensure that each uptime accumulator value that was passed in as an argument changes by the correct amount.
				for uptimeIndex := range uptimeAccs {
					expectedValue := initUptimeAccumValues[uptimeIndex].Add(expectedUptimeDeltas[uptimeIndex]...)
					s.Require().Equal(expectedValue, uptimeAccs[uptimeIndex].GetValue())
				}
			})
		}
	})
}

// Note: we test that incentive records are properly deducted by emissions in `TestUpdateUptimeAccumulatorsToNow` above.
// This test aims to cover the behavior of a series of state read/writes relating to incentive records.
// Since these are lower level functions, we expect that validation logic for authorized uptimes is done at a higher level (see `TestCreateIncentive` tests).
func (s *KeeperTestSuite) TestIncentiveRecordsSetAndGet() {
	s.SetupTest()
	clKeeper := s.App.ConcentratedLiquidityKeeper
	s.Ctx = s.Ctx.WithBlockTime(defaultStartTime)
	emptyIncentiveRecords := []types.IncentiveRecord{}

	// Set up two test pool
	clPoolOne := s.PrepareConcentratedPool()
	clPoolTwo := s.PrepareConcentratedPool()
	ensurePoolTwoRecordsEmpty := func() {
		poolTwoRecords := s.getAllIncentiveRecordsForPool(clPoolTwo.GetId())
		s.Require().Equal(emptyIncentiveRecords, poolTwoRecords)
	}

	// Ensure both pools start with no incentive records
	poolOneRecords := s.getAllIncentiveRecordsForPool(clPoolOne.GetId())
	s.Require().Equal(emptyIncentiveRecords, poolOneRecords)
	ensurePoolTwoRecordsEmpty()

	// Ensure setting and getting a single record works with single Get and GetAll
	err := clKeeper.SetIncentiveRecord(s.Ctx, incentiveRecordOne)
	s.Require().NoError(err)
	poolOneRecord, err := clKeeper.GetIncentiveRecord(s.Ctx, clPoolOne.GetId(), incentiveRecordOne.MinUptime, incentiveRecordOne.IncentiveId)
	s.Require().NoError(err)
	s.Require().Equal(incentiveRecordOne, poolOneRecord)
	allRecordsPoolOne := s.getAllIncentiveRecordsForPool(clPoolOne.GetId())
	s.Require().Equal([]types.IncentiveRecord{incentiveRecordOne}, allRecordsPoolOne)

	// Ensure records for other pool remain unchanged
	poolTwoRecord, err := clKeeper.GetIncentiveRecord(s.Ctx, clPoolTwo.GetId(), incentiveRecordOne.MinUptime, incentiveRecordOne.IncentiveId)
	s.Require().ErrorIs(err, types.IncentiveRecordNotFoundError{PoolId: clPoolTwo.GetId(), MinUptime: incentiveRecordOne.MinUptime, IncentiveRecordId: incentiveRecordOne.IncentiveId})
	s.Require().Equal(types.IncentiveRecord{}, poolTwoRecord)
	ensurePoolTwoRecordsEmpty()

	// Ensure directly setting additional records don't overwrite previous ones
	err = clKeeper.SetIncentiveRecord(s.Ctx, incentiveRecordTwo)
	s.Require().NoError(err)
	poolOneRecord, err = clKeeper.GetIncentiveRecord(s.Ctx, clPoolOne.GetId(), incentiveRecordTwo.MinUptime, incentiveRecordTwo.IncentiveId)
	s.Require().NoError(err)
	s.Require().Equal(incentiveRecordTwo, poolOneRecord)
	allRecordsPoolOne = s.getAllIncentiveRecordsForPool(clPoolOne.GetId())
	s.Require().Equal([]types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo}, allRecordsPoolOne)

	// Ensure that directly setting the same record, completely overwrites the previous one
	// with the same id.
	err = clKeeper.SetIncentiveRecord(s.Ctx, incentiveRecordTwo)
	s.Require().NoError(err)
	allRecordsPoolOne = s.getAllIncentiveRecordsForPool(clPoolOne.GetId())
	s.Require().Equal([]types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo}, allRecordsPoolOne)

	// Ensure setting multiple records through helper functions as expected
	// Note that we also pass in an empty incentive record, which we expect to be cleared out while being set
	err = clKeeper.SetMultipleIncentiveRecords(s.Ctx, []types.IncentiveRecord{incentiveRecordThree, incentiveRecordFour, emptyIncentiveRecord})
	s.Require().NoError(err)

	// Note: we expect the records to be retrieved in lexicographic order by denom and for the empty record to be cleared
	allRecordsPoolOne = s.getAllIncentiveRecordsForPool(clPoolOne.GetId())
	s.Require().Equal([]types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo, incentiveRecordThree, incentiveRecordFour}, allRecordsPoolOne)

	// Finally, we ensure the second pool remains unaffected
	ensurePoolTwoRecordsEmpty()
}

func (s *KeeperTestSuite) TestGetInitialUptimeGrowthOppositeDirectionOfLastTraversalForTick() {
	expectedUptimes := getExpectedUptimes()

	type getInitialUptimeGrowthOppositeDirectionOfLastTraversalForTick struct {
		tick                            int64
		expectedUptimeAccumulatorValues []sdk.DecCoins
	}
	tests := map[string]getInitialUptimeGrowthOppositeDirectionOfLastTraversalForTick{
		"uptime growth for tick <= currentTick": {
			tick:                            -2,
			expectedUptimeAccumulatorValues: expectedUptimes.hundredTokensMultiDenom,
		},
		"uptime growth for tick > currentTick": {
			tick:                            1,
			expectedUptimeAccumulatorValues: expectedUptimes.emptyExpectedAccumValues,
		},
	}

	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			tc := tc

			s.Run(name, func() {
				s.SetupTest()
				clKeeper := s.App.ConcentratedLiquidityKeeper

				pool := s.PrepareConcentratedPool()

				poolUptimeAccumulators, err := clKeeper.GetUptimeAccumulators(s.Ctx, pool.GetId())
				s.Require().NoError(err)

				for uptimeId, uptimeAccum := range poolUptimeAccumulators {
					uptimeAccum.AddToAccumulator(expectedUptimes.hundredTokensMultiDenom[uptimeId])
				}

				_, err = clKeeper.GetUptimeAccumulators(s.Ctx, pool.GetId())
				s.Require().NoError(err)

				val, err := clKeeper.GetInitialUptimeGrowthOppositeDirectionOfLastTraversalForTick(s.Ctx, pool, tc.tick)
				s.Require().NoError(err)
				s.Require().Equal(val, tc.expectedUptimeAccumulatorValues)
			})
		}
	})
}

// Test uptime growth inside and outside range.
func (s *KeeperTestSuite) TestGetUptimeGrowthRange() {
	defaultInitialLiquidity := osmomath.OneDec()
	uptimeHelper := getExpectedUptimes()

	type uptimeGrowthTest struct {
		lowerTick                    int64
		upperTick                    int64
		currentTick                  int64
		lowerTickUptimeGrowthOutside []sdk.DecCoins
		upperTickUptimeGrowthOutside []sdk.DecCoins
		globalUptimeGrowth           []sdk.DecCoins

		expectedUptimeGrowthInside  []sdk.DecCoins
		expectedUptimeGrowthOutside []sdk.DecCoins
	}

	tests := map[string]uptimeGrowthTest{
		// current tick above range
		"current tick > upper tick, nonzero uptime growth inside": {
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  2,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is above range, we expect upper - lower
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			// Since current tick is above range, we expect global - (upper - lower)
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
		},
		"current tick > upper tick, nonzero uptime growth inside (wider range)": {
			lowerTick:                    12444,
			upperTick:                    15013,
			currentTick:                  50320,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is above range, we expect upper - lower
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			// Since current tick is above range, we expect global - (upper - lower)
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
		},
		"current tick > upper tick, zero uptime growth inside (nonempty trackers)": {
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  2,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is above range, we expect upper - lower
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.threeHundredTokensMultiDenom,
		},
		"current tick > upper tick, zero uptime growth inside (empty trackers)": {
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  2,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since current tick is above range, we expect upper - lower
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
		},
		"current tick > upper tick, zero uptime growth inside with extraneous uptime growth": {
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  2,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.twoHundredTokensMultiDenom,

			// Since current tick is above range, we expect upper - lower
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
		},

		// current tick within range

		"upper tick > current tick > lower tick, nonzero uptime growth inside": {
			lowerTick:                    0,
			upperTick:                    2,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is within range, we expect global - upper - lower
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			// Since current tick is within range, we expect global - (global - upper - lower)
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
		},
		"upper tick > current tick > lower tick, nonzero uptime growth inside (wider range)": {
			lowerTick:                    -19753,
			upperTick:                    8921,
			currentTick:                  -97,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is within range, we expect global - upper - lower
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			// Since current tick is within range, we expect global - (global - upper - lower)
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
		},
		"upper tick > current tick > lower tick, zero uptime growth inside (nonempty trackers)": {
			lowerTick:                    0,
			upperTick:                    2,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.twoHundredTokensMultiDenom,

			// Since current tick is within range, we expect global - upper - lower
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
		},
		"upper tick > current tick > lower tick, zero uptime growth inside (empty trackers)": {
			lowerTick:                    0,
			upperTick:                    2,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since current tick is within range, we expect global - upper - lower
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
		},

		// current tick below range

		"current tick < lower tick, nonzero uptime growth inside": {
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  -1,
			lowerTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is below range, we expect lower - upper
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			// Since current tick is below range, we expect global - (lower - upper)
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
		},
		"current tick < lower tick, nonzero uptime growth inside (wider range)": {
			lowerTick:                    328,
			upperTick:                    726,
			currentTick:                  189,
			lowerTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is below range, we expect lower - upper
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			// Since current tick is below range, we expect global - (lower - upper)
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
		},
		"current tick < lower tick, zero uptime growth inside (nonempty trackers)": {
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  -1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is below range, we expect lower - upper
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.threeHundredTokensMultiDenom,
		},
		"current tick < lower tick, zero uptime growth inside (empty trackers)": {
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  -1,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since current tick is below range, we expect lower - upper
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
		},

		// current tick on range boundary

		"current tick = lower tick, nonzero uptime growth inside": {
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  0,
			lowerTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.sixHundredTokensMultiDenom,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being within the range (global - upper - lower)
			expectedUptimeGrowthInside: uptimeHelper.threeHundredTokensMultiDenom,
			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being within the range (global - (global - upper - lower))
			expectedUptimeGrowthOutside: uptimeHelper.threeHundredTokensMultiDenom,
		},
		"current tick = lower tick, zero uptime growth inside (nonempty trackers)": {
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  0,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.sixHundredTokensMultiDenom,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being within the range (global - upper - lower)
			expectedUptimeGrowthInside: uptimeHelper.fourHundredTokensMultiDenom,
			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being within the range (global - (global - upper - lower))
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
		},
		"current tick = lower tick, zero uptime growth inside (empty trackers)": {
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  0,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being within the range (global - upper - lower)
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
		},
		"current tick = upper tick, nonzero uptime growth inside": {
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.fourHundredTokensMultiDenom,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being above the range (upper - lower)
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being above the range (global - (upper - lower))
			expectedUptimeGrowthOutside: uptimeHelper.threeHundredTokensMultiDenom,
		},
		"current tick = upper tick, zero uptime growth inside (nonempty trackers)": {
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.twoHundredTokensMultiDenom,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being above the range (upper - lower)
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
		},
		"current tick = upper tick, zero uptime growth inside (empty trackers)": {
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being above the range (upper - lower)
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
		},
	}

	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			s.Run(name, func() {
				s.SetupTest()

				pool := s.PrepareConcentratedPool()

				// Note: we scale all these values up as addToUptimeAccums(...) does the same for the global uptime accums.
				tc.lowerTickUptimeGrowthOutside = s.scaleUptimeAccumulators(tc.lowerTickUptimeGrowthOutside)
				tc.upperTickUptimeGrowthOutside = s.scaleUptimeAccumulators(tc.upperTickUptimeGrowthOutside)
				tc.expectedUptimeGrowthInside = s.scaleUptimeAccumulators(tc.expectedUptimeGrowthInside)
				tc.expectedUptimeGrowthOutside = s.scaleUptimeAccumulators(tc.expectedUptimeGrowthOutside)

				// Update global uptime accums
				err := addToUptimeAccums(s.Ctx, pool.GetId(), s.App.ConcentratedLiquidityKeeper, tc.globalUptimeGrowth)
				s.Require().NoError(err)

				// Update tick-level uptime trackers
				s.initializeTick(s.Ctx, tc.lowerTick, defaultInitialLiquidity, cl.EmptyCoins, wrapUptimeTrackers(tc.lowerTickUptimeGrowthOutside), true)
				s.initializeTick(s.Ctx, tc.upperTick, defaultInitialLiquidity, cl.EmptyCoins, wrapUptimeTrackers(tc.upperTickUptimeGrowthOutside), false)
				pool.SetCurrentTick(tc.currentTick)
				err = s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, pool)
				s.Require().NoError(err)

				// system under test
				uptimeGrowthInside, err := s.App.ConcentratedLiquidityKeeper.GetUptimeGrowthInsideRange(s.Ctx, pool.GetId(), tc.lowerTick, tc.upperTick)
				s.Require().NoError(err)

				// check if returned uptime growth inside has correct value
				s.RequireDecCoinsSlice(tc.expectedUptimeGrowthInside, uptimeGrowthInside)

				uptimeGrowthOutside, err := s.App.ConcentratedLiquidityKeeper.GetUptimeGrowthOutsideRange(s.Ctx, pool.GetId(), tc.lowerTick, tc.upperTick)
				s.Require().NoError(err)

				// check if returned uptime growth inside has correct value
				s.RequireDecCoinsSlice(tc.expectedUptimeGrowthOutside, uptimeGrowthOutside)
			})
		}
	})
}

func (s *KeeperTestSuite) TestGetUptimeGrowthErrors() {
	_, err := s.Clk.GetUptimeGrowthInsideRange(s.Ctx, defaultPoolId, 0, 0)
	s.Require().Error(err)
	_, err = s.Clk.GetUptimeGrowthOutsideRange(s.Ctx, defaultPoolId, 0, 0)
	s.Require().Error(err)
}

// TestInitOrUpdatePositionUptimeAccumulators tests the subset of position update logic related to checkpointing uptime accumulators.
// Since all supported uptime accumulators are checkpointed regardless of which are authorized, we test on all of them.
func (s *KeeperTestSuite) TestInitOrUpdatePositionUptimeAccumulators() {
	uptimeHelper := getExpectedUptimes()
	type tick struct {
		tickIndex      int64
		uptimeTrackers []model.UptimeTracker
	}

	tests := map[string]struct {
		positionLiquidity osmomath.Dec

		lowerTick               tick
		upperTick               tick
		positionId              uint64
		currentTickIndex        int64
		globalUptimeAccumValues []sdk.DecCoins

		// For testing updates on existing liquidity
		existingPosition  bool
		newLowerTick      tick
		newUpperTick      tick
		addToGlobalAccums []sdk.DecCoins

		expectedInitAccumValue   []sdk.DecCoins
		expectedUnclaimedRewards []sdk.DecCoins
		expectedErr              error
	}{
		// New position tests

		"(lower < curr < upper) nonzero uptime trackers": {
			positionLiquidity: DefaultLiquidityAmt,
			// Note: that we scale uptime tracker values up as addToUptimeAccums(...) does the same for the global uptime accums.
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.hundredTokensMultiDenom)),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.hundredTokensMultiDenom)),
			},
			positionId:               DefaultPositionId,
			currentTickIndex:         0,
			globalUptimeAccumValues:  uptimeHelper.threeHundredTokensMultiDenom,
			expectedInitAccumValue:   uptimeHelper.hundredTokensMultiDenom,
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
		"(lower < curr < upper) non-zero uptime trackers (position already existing)": {
			positionLiquidity: DefaultLiquidityAmt,
			// Note: that we scale uptime tracker values up as addToUptimeAccums(...) does the same for the global uptime accums.
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.hundredTokensMultiDenom)),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.hundredTokensMultiDenom)),
			},
			existingPosition:  true,
			addToGlobalAccums: uptimeHelper.threeHundredTokensMultiDenom,
			positionId:        DefaultPositionId,
			currentTickIndex:  0,

			// The global accum value here is arbitrarily chosen to determine what we initialize the global accumulators to.
			globalUptimeAccumValues: uptimeHelper.threeHundredTokensMultiDenom,
			// We start with `hundredTokensMultiDenom` in the tick trackers (which is in range), and add `threeHundredTokensMultiDenom` to the global accumulators.
			// This gives us an expected init value of `fourHundredTokensMultiDenom`.
			expectedInitAccumValue: uptimeHelper.fourHundredTokensMultiDenom,
			// The unclaimed rewards matches the amount we added to the global accumulators (since the position already existed and was in range).
			expectedUnclaimedRewards: uptimeHelper.threeHundredTokensMultiDenom,
		},
		"(lower < upper < curr) nonzero uptime trackers": {
			positionLiquidity: DefaultLiquidityAmt,
			// Note: that we scale uptime tracker values up as addToUptimeAccums(...) does the same for the global uptime accums.
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.hundredTokensMultiDenom)),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.threeHundredTokensMultiDenom)),
			},
			positionId:               DefaultPositionId,
			currentTickIndex:         51,
			globalUptimeAccumValues:  uptimeHelper.fourHundredTokensMultiDenom,
			expectedInitAccumValue:   uptimeHelper.twoHundredTokensMultiDenom,
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
		"(lower < upper < curr) non-zero uptime trackers (position already existing)": {
			positionLiquidity: DefaultLiquidityAmt,
			// Note: that we scale uptime tracker values up as addToUptimeAccums(...) does the same for the global uptime accums.
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.hundredTokensMultiDenom)),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.threeHundredTokensMultiDenom)),
			},
			existingPosition:  true,
			addToGlobalAccums: uptimeHelper.threeHundredTokensMultiDenom,
			positionId:        DefaultPositionId,
			currentTickIndex:  51,

			// The global accum value here is arbitrarily chosen to determine what we initialize the global accumulators to.
			globalUptimeAccumValues: uptimeHelper.fourHundredTokensMultiDenom,
			// The difference between the lower and upper tick's uptime trackers is `twoHundredTokensMultiDenom`.
			// The amount we added to the global accumulators is `threeHundredTokensMultiDenom`, but doesn't count towards the init value since the position was out of range.
			expectedInitAccumValue: uptimeHelper.twoHundredTokensMultiDenom,
			// The unclaimed rewards here is still empty despite having a pre-existing position, because the position has been out of range for the entire time.
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
		"(curr < lower < upper) nonzero uptime trackers": {
			positionLiquidity: DefaultLiquidityAmt,
			// Note: that we scale uptime tracker values up as addToUptimeAccums(...) does the same for the global uptime accums.
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.threeHundredTokensMultiDenom)),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.hundredTokensMultiDenom)),
			},
			positionId:               DefaultPositionId,
			currentTickIndex:         -51,
			globalUptimeAccumValues:  uptimeHelper.fourHundredTokensMultiDenom,
			expectedInitAccumValue:   uptimeHelper.twoHundredTokensMultiDenom,
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
		"(curr < lower < upper) nonzero uptime trackers (position already existing)": {
			positionLiquidity: DefaultLiquidityAmt,
			// Note: that we scale uptime tracker values up as addToUptimeAccums(...) does the same for the global uptime accums.
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.threeHundredTokensMultiDenom)),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.hundredTokensMultiDenom)),
			},
			existingPosition:  true,
			addToGlobalAccums: uptimeHelper.threeHundredTokensMultiDenom,
			positionId:        DefaultPositionId,
			currentTickIndex:  -51,

			// The global accum value here is arbitrarily chosen to determine what we initialize the global accumulators to.
			globalUptimeAccumValues: uptimeHelper.fourHundredTokensMultiDenom,
			// The difference between the lower and upper tick's uptime trackers is `twoHundredTokensMultiDenom`.
			// The amount we added to the global accumulators is `threeHundredTokensMultiDenom`, but doesn't count towards the init value since the position was out of range.
			expectedInitAccumValue: uptimeHelper.twoHundredTokensMultiDenom,
			// The unclaimed rewards here is still empty despite having a pre-existing position, because the position has been out of range for the entire time.
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
		"(lower < curr < upper) nonzero and variable uptime trackers": {
			positionLiquidity: DefaultLiquidityAmt,
			// Note: that we scale uptime tracker values up as addToUptimeAccums(...) does the same for the global uptime accums.
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.varyingTokensMultiDenom)),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.hundredTokensMultiDenom)),
			},
			positionId:       DefaultPositionId,
			currentTickIndex: 0,

			// We set up the global accum values such that the growth inside is equal to 100 of each denom
			// for each uptime tracker. Let the uptime growth inside (UGI) = 100
			globalUptimeAccumValues: []sdk.DecCoins{
				sdk.NewDecCoins(
					// 100 + 100 + UGI = 300
					sdk.NewDecCoin("bar", osmomath.NewInt(300)),
					// 100 + 100 + UGI = 300
					sdk.NewDecCoin("foo", osmomath.NewInt(300)),
				),
				sdk.NewDecCoins(
					// 100 + 103 + UGI = 303
					sdk.NewDecCoin("bar", osmomath.NewInt(303)),
					// 100 + 101 + UGI = 301
					sdk.NewDecCoin("foo", osmomath.NewInt(301)),
				),
				sdk.NewDecCoins(
					// 100 + 106 + UGI = 306
					sdk.NewDecCoin("bar", osmomath.NewInt(306)),
					// 100 + 102 + UGI = 302
					sdk.NewDecCoin("foo", osmomath.NewInt(302)),
				),
				sdk.NewDecCoins(
					// 100 + 109 + UGI = 309
					sdk.NewDecCoin("bar", osmomath.NewInt(309)),
					// 100 + 103 + UGI = 303
					sdk.NewDecCoin("foo", osmomath.NewInt(303)),
				),
				sdk.NewDecCoins(
					// 100 + 112 + UGI = 312
					sdk.NewDecCoin("bar", osmomath.NewInt(312)),
					// 100 + 104 + UGI = 304
					sdk.NewDecCoin("foo", osmomath.NewInt(304)),
				),
				sdk.NewDecCoins(
					// 100 + 115 + UGI = 315
					sdk.NewDecCoin("bar", osmomath.NewInt(315)),
					// 100 + 105 + UGI = 305
					sdk.NewDecCoin("foo", osmomath.NewInt(305)),
				),
			},
			// Equal to 100 of foo and bar in each uptime tracker (UGI)
			expectedInitAccumValue:   uptimeHelper.hundredTokensMultiDenom,
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
		"error: negative liquidity for first position": {
			positionLiquidity: DefaultLiquidityAmt.Neg(),
			// Note: that we scale uptime tracker values up as addToUptimeAccums(...) does the same for the global uptime accums.
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.hundredTokensMultiDenom)),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(uptimeHelper.hundredTokensMultiDenom)),
			},
			positionId:              DefaultPositionId,
			currentTickIndex:        0,
			globalUptimeAccumValues: uptimeHelper.threeHundredTokensMultiDenom,

			expectedErr: types.NonPositiveLiquidityForNewPositionError{PositionId: DefaultPositionId, LiquidityDelta: DefaultLiquidityAmt.Neg()},
		},
	}

	s.runMultipleAuthorizedUptimes(func() {
		for name, test := range tests {
			s.Run(name, func() {
				// --- Setup ---

				// Init suite for each test.
				s.SetupTest()

				// Note: that we scale all these values up as addToUptimeAccums(...) does the same for the global uptime accums.
				test.expectedInitAccumValue = s.scaleUptimeAccumulators(test.expectedInitAccumValue)
				test.expectedUnclaimedRewards = s.scaleUptimeAccumulators(test.expectedUnclaimedRewards)

				// Set blocktime to fixed UTC value for consistency
				s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

				clPool := s.PrepareConcentratedPool()

				// Initialize lower, upper, and current ticksosmomath.ZeroDec
				s.initializeTick(s.Ctx, test.lowerTick.tickIndex, osmomath.ZeroDec(), cl.EmptyCoins, test.lowerTick.uptimeTrackers, true)
				s.initializeTick(s.Ctx, test.upperTick.tickIndex, osmomath.ZeroDec(), cl.EmptyCoins, test.upperTick.uptimeTrackers, false)
				clPool.SetCurrentTick(test.currentTickIndex)
				err := s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, clPool)
				s.Require().NoError(err)

				// Initialize global uptime accums
				err = addToUptimeAccums(s.Ctx, clPool.GetId(), s.App.ConcentratedLiquidityKeeper, test.globalUptimeAccumValues)
				s.Require().NoError(err)

				// If applicable, set up existing position and update ticks & global accums
				if test.existingPosition {
					err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePositionUptimeAccumulators(s.Ctx, clPool.GetId(), test.positionLiquidity, test.lowerTick.tickIndex, test.upperTick.tickIndex, test.positionLiquidity, DefaultPositionId)
					s.Require().NoError(err)
					err = s.App.ConcentratedLiquidityKeeper.SetPosition(s.Ctx, clPool.GetId(), s.TestAccs[0], test.lowerTick.tickIndex, test.upperTick.tickIndex, DefaultJoinTime, test.positionLiquidity, DefaultPositionId, DefaultUnderlyingLockId)
					s.Require().NoError(err)
					s.initializeTick(s.Ctx, test.newLowerTick.tickIndex, osmomath.ZeroDec(), cl.EmptyCoins, test.newLowerTick.uptimeTrackers, true)
					s.initializeTick(s.Ctx, test.newUpperTick.tickIndex, osmomath.ZeroDec(), cl.EmptyCoins, test.newUpperTick.uptimeTrackers, false)
					clPool.SetCurrentTick(test.currentTickIndex)
					err = s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, clPool)
					s.Require().NoError(err)

					err = addToUptimeAccums(s.Ctx, clPool.GetId(), s.App.ConcentratedLiquidityKeeper, test.addToGlobalAccums)
					s.Require().NoError(err)
				}

				// --- System under test ---

				err = s.App.ConcentratedLiquidityKeeper.InitOrUpdatePositionUptimeAccumulators(s.Ctx, clPool.GetId(), test.positionLiquidity, test.lowerTick.tickIndex, test.upperTick.tickIndex, test.positionLiquidity, DefaultPositionId)

				// --- Error catching ---

				if test.expectedErr != nil {
					s.Require().ErrorContains(err, test.expectedErr.Error())
					return
				}

				// --- Non error case checks ---

				s.Require().NoError(err)

				// Pre-compute variables for readability
				positionName := string(types.KeyPositionId(test.positionId))
				uptimeAccums, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, clPool.GetId())
				s.Require().NoError(err)

				// Ensure records are properly updated for each supported uptime
				for uptimeIndex := range types.SupportedUptimes {
					recordExists := uptimeAccums[uptimeIndex].HasPosition(positionName)
					s.Require().True(recordExists)

					// Ensure position's record has correct values
					positionRecord, err := accum.GetPosition(uptimeAccums[uptimeIndex], positionName)
					s.Require().NoError(err)

					s.Require().Equal(test.expectedInitAccumValue[uptimeIndex], positionRecord.AccumValuePerShare)

					if test.existingPosition {
						s.Require().Equal(osmomath.NewDec(2).Mul(test.positionLiquidity), positionRecord.NumShares)
					} else {
						s.Require().Equal(test.positionLiquidity, positionRecord.NumShares)
					}

					// Note that the rewards only apply to the initial shares, not the new ones
					s.Require().Equal(test.expectedUnclaimedRewards[uptimeIndex].MulDec(test.positionLiquidity), positionRecord.UnclaimedRewardsTotal)
				}
			})
		}
	})
}

// TestQueryAndCollectIncentives tests that incentive queries are correct by collecting incentives for a position and comparing the results to the query.
func (s *KeeperTestSuite) TestQueryAndCollectIncentives() {
	ownerWithValidPosition := s.TestAccs[0]
	uptimeHelper := getExpectedUptimes()
	oneDay := time.Hour * 24
	twoWeeks := 14 * time.Hour * 24
	defaultJoinTime := DefaultJoinTime

	type positionParameters struct {
		owner       sdk.AccAddress
		lowerTick   int64
		upperTick   int64
		liquidity   osmomath.Dec
		joinTime    time.Time
		collectTime time.Time
		positionId  uint64
	}

	default0To2PosParam := positionParameters{
		owner:       ownerWithValidPosition,
		lowerTick:   0,
		upperTick:   2,
		liquidity:   DefaultLiquidityAmt,
		joinTime:    defaultJoinTime,
		positionId:  DefaultPositionId,
		collectTime: defaultJoinTime.Add(100),
	}
	default1To2PosParam := default0To2PosParam
	default1To2PosParam.lowerTick = 1

	tests := map[string]struct {
		// setup parameters
		existingAccumLiquidity   []osmomath.Dec
		addedUptimeGrowthInside  []sdk.DecCoins
		addedUptimeGrowthOutside []sdk.DecCoins
		currentTick              int64
		numPositions             int

		// inputs parameters
		positionParams positionParameters
		timeInPosition time.Duration

		// expectations
		expectedIncentivesClaimed   sdk.Coins
		expectedForfeitedIncentives sdk.Coins
		expectedError               error
	}{
		// ---Cases for lowerTick < currentTick < upperTick---

		"(lower < curr < upper) no uptime growth inside or outside range, 1D time in position": {
			currentTick:                 1,
			positionParams:              default0To2PosParam,
			numPositions:                1,
			timeInPosition:              oneDay,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth outside range but not inside, 1D time in position": {
			currentTick:              1,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default0To2PosParam,
			numPositions:             1,
			timeInPosition:           oneDay,
			// Since there was no growth inside the range, we expect no incentives to be claimed
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth inside range but not outside, 1D time in position": {
			currentTick:             1,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:          default0To2PosParam,
			numPositions:            1,
			timeInPosition:          oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)...),
		},
		"(lower < curr < upper) uptime growth both inside and outside range, 1D time in position": {
			currentTick:              1,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default0To2PosParam,
			numPositions:             1,
			timeInPosition:           oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for. At the same time, growth outside does not affect the current position's incentive rewards.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)...),
		},
		"(lower < curr < upper) no uptime growth inside or outside range, 2W time in position": {
			currentTick:                 1,
			positionParams:              default0To2PosParam,
			numPositions:                1,
			timeInPosition:              twoWeeks,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth outside range but not inside, 2W time in position": {
			currentTick:              1,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default0To2PosParam,
			numPositions:             1,
			timeInPosition:           twoWeeks,
			// Since there was no growth inside the range, we expect no incentives to be claimed
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth inside range but not outside, 2W time in position": {
			currentTick:             1,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:          default0To2PosParam,
			numPositions:            1,
			timeInPosition:          twoWeeks,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth both inside and outside range, 2W time in position": {
			currentTick:              1,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default0To2PosParam,
			numPositions:             1,
			timeInPosition:           twoWeeks,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for. At the same time, growth outside does not affect the current position's incentive rewards.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < curr < upper) no uptime growth inside or outside range, no time in position": {
			currentTick:                 1,
			positionParams:              default0To2PosParam,
			numPositions:                1,
			timeInPosition:              0,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth outside range but not inside, no time in position": {
			currentTick:                 1,
			addedUptimeGrowthOutside:    uptimeHelper.hundredTokensMultiDenom,
			positionParams:              default0To2PosParam,
			numPositions:                1,
			timeInPosition:              0,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth inside range but not outside, no time in position": {
			currentTick:                 1,
			addedUptimeGrowthInside:     uptimeHelper.hundredTokensMultiDenom,
			positionParams:              default0To2PosParam,
			numPositions:                1,
			timeInPosition:              0,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier),
		},
		"(lower < curr < upper) uptime growth both inside and outside range, no time in position": {
			currentTick:                 1,
			addedUptimeGrowthInside:     uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside:    uptimeHelper.hundredTokensMultiDenom,
			positionParams:              default0To2PosParam,
			numPositions:                1,
			timeInPosition:              0,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier),
		},

		// ---Cases for currentTick < lowerTick < upperTick---

		"(curr < lower < upper) no uptime growth inside or outside range, 1D time in position": {
			currentTick:                 0,
			positionParams:              default1To2PosParam,
			numPositions:                1,
			timeInPosition:              oneDay,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth outside range but not inside, 1D time in position": {
			currentTick:              0,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default1To2PosParam,
			numPositions:             1,
			timeInPosition:           oneDay,
			// Since there was no growth inside the range, we expect no incentives to be claimed
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth inside range but not outside, 1D time in position": {
			currentTick:             0,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:          default1To2PosParam,
			numPositions:            1,
			timeInPosition:          oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)...),
		},
		"(curr < lower < upper) uptime growth both inside and outside range, 1D time in position": {
			currentTick:              0,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default1To2PosParam,
			numPositions:             1,
			timeInPosition:           oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for. At the same time, growth outside does not affect the current position's incentive rewards.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)...),
		},
		"(curr < lower < upper) no uptime growth inside or outside range, 2W time in position": {
			currentTick:                 0,
			positionParams:              default1To2PosParam,
			numPositions:                1,
			timeInPosition:              twoWeeks,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth outside range but not inside, 2W time in position": {
			currentTick:              0,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default1To2PosParam,
			numPositions:             1,
			timeInPosition:           twoWeeks,
			// Since there was no growth inside the range, we expect no incentives to be claimed
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth inside range but not outside, 2W time in position": {
			currentTick:             0,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:          default1To2PosParam,
			numPositions:            1,
			timeInPosition:          twoWeeks,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth both inside and outside range, 2W time in position": {
			currentTick:              0,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default1To2PosParam,
			numPositions:             1,
			timeInPosition:           twoWeeks,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for. At the same time, growth outside does not affect the current position's incentive rewards.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(curr < lower < upper) no uptime growth inside or outside range, no time in position": {
			currentTick:                 0,
			positionParams:              default1To2PosParam,
			numPositions:                1,
			timeInPosition:              0,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth outside range but not inside, no time in position": {
			currentTick:                 0,
			addedUptimeGrowthOutside:    uptimeHelper.hundredTokensMultiDenom,
			positionParams:              default1To2PosParam,
			numPositions:                1,
			timeInPosition:              0,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth inside range but not outside, no time in position": {
			currentTick:                 0,
			addedUptimeGrowthInside:     uptimeHelper.hundredTokensMultiDenom,
			positionParams:              default1To2PosParam,
			numPositions:                1,
			timeInPosition:              0,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier),
		},
		"(curr < lower < upper) uptime growth both inside and outside range, no time in position": {
			currentTick:                 0,
			addedUptimeGrowthInside:     uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside:    uptimeHelper.hundredTokensMultiDenom,
			positionParams:              default1To2PosParam,
			numPositions:                1,
			timeInPosition:              0,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier),
		},

		// ---Cases for lowerTick < upperTick < currentTick---

		"(lower < upper < curr) no uptime growth inside or outside range, 1D time in position": {
			currentTick:                 3,
			positionParams:              default1To2PosParam,
			numPositions:                1,
			timeInPosition:              oneDay,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth outside range but not inside, 1D time in position": {
			currentTick:              3,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default1To2PosParam,
			numPositions:             1,
			timeInPosition:           oneDay,
			// Since there was no growth inside the range, we expect no incentives to be claimed
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth inside range but not outside, 1D time in position": {
			currentTick:             3,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:          default1To2PosParam,
			numPositions:            1,
			timeInPosition:          oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)...),
		},
		"(lower < upper < curr) uptime growth both inside and outside range, 1D time in position": {
			currentTick:              3,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default1To2PosParam,
			numPositions:             1,
			timeInPosition:           oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for. At the same time, growth outside does not affect the current position's incentive rewards.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)...),
		},
		"(lower < upper < curr) no uptime growth inside or outside range, 1W time in position": {
			currentTick:                 3,
			positionParams:              default1To2PosParam,
			numPositions:                1,
			timeInPosition:              twoWeeks,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth outside range but not inside, 1W time in position": {
			currentTick:              3,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default1To2PosParam,
			numPositions:             1,
			timeInPosition:           twoWeeks,
			// Since there was no growth inside the range, we expect no incentives to be claimed
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth inside range but not outside, 1W time in position": {
			currentTick:             3,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:          default1To2PosParam,
			numPositions:            1,
			timeInPosition:          twoWeeks,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth both inside and outside range, 1W time in position": {
			currentTick:              3,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default1To2PosParam,
			numPositions:             1,
			timeInPosition:           twoWeeks,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < upper < curr) no uptime growth inside or outside range, no time in position": {
			currentTick:                 3,
			positionParams:              default1To2PosParam,
			numPositions:                1,
			timeInPosition:              0,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth outside range but not inside, no time in position": {
			currentTick:                 3,
			addedUptimeGrowthOutside:    uptimeHelper.hundredTokensMultiDenom,
			positionParams:              default1To2PosParam,
			numPositions:                1,
			timeInPosition:              0,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth inside range but not outside, no time in position": {
			currentTick:                 3,
			addedUptimeGrowthInside:     uptimeHelper.hundredTokensMultiDenom,
			positionParams:              default1To2PosParam,
			numPositions:                1,
			timeInPosition:              0,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier),
		},
		"(lower < upper < curr) uptime growth both inside and outside range, no time in position": {
			currentTick:                 3,
			addedUptimeGrowthInside:     uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside:    uptimeHelper.hundredTokensMultiDenom,
			positionParams:              default1To2PosParam,
			numPositions:                1,
			timeInPosition:              0,
			expectedIncentivesClaimed:   sdk.Coins(nil),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier),
		},

		// Edge case tests

		"(curr = lower) uptime growth both inside and outside range, 1D time in position": {
			currentTick:              0,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default0To2PosParam,
			numPositions:             1,
			timeInPosition:           oneDay,
			// We expect this case to behave like (lower < curr < upper)
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)...),
		},
		"(curr = upper) uptime growth both inside and outside range, 1D time in position": {
			currentTick:              2,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default1To2PosParam,
			numPositions:             1,
			timeInPosition:           oneDay,
			// We expect this case to behave like (lower < upper < curr)
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)...),
		},
		"other liquidity on uptime accums: (lower < curr < upper) uptime growth both inside and outside range, 1D time in position": {
			currentTick: 1,
			existingAccumLiquidity: []osmomath.Dec{
				osmomath.NewDec(99900123432),
				osmomath.NewDec(18942),
				osmomath.NewDec(0),
				osmomath.NewDec(9981),
				osmomath.NewDec(1),
				osmomath.NewDec(778212931834),
			},
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default0To2PosParam,
			numPositions:             1,
			timeInPosition:           oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)...),
		},
		"multiple positions in same range: (lower < curr < upper) uptime growth both inside and outside range, 1D time in position": {
			currentTick: 1,
			existingAccumLiquidity: []osmomath.Dec{
				osmomath.NewDec(99900123432),
				osmomath.NewDec(18942),
				osmomath.NewDec(0),
				osmomath.NewDec(9981),
				osmomath.NewDec(1),
				osmomath.NewDec(778212931834),
			},
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default0To2PosParam,
			numPositions:             3,
			timeInPosition:           oneDay,
			// Since we introduced positionIDs, despite these position having the same range and pool, only
			// the position ID being claimed will be considered for the claim.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)...),
		},

		// Error catching

		"position does not exist": {
			currentTick: 1,

			numPositions: 0,

			expectedIncentivesClaimed:   sdk.Coins{},
			expectedForfeitedIncentives: sdk.Coins{},
			expectedError:               types.PositionIdNotFoundError{PositionId: DefaultPositionId},
		},
	}

	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			s.Run(name, func() {
				tc := tc
				s.SetupTest()

				// We fix join time so tests are deterministic
				s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)

				validPool := s.PrepareConcentratedPool()
				validPoolId := validPool.GetId()

				// Fund the incentives address with amount we intend to claim and forfeit.
				s.FundAcc(validPool.GetIncentivesAddress(), tc.expectedIncentivesClaimed.Add(tc.expectedForfeitedIncentives...))

				clKeeper := s.App.ConcentratedLiquidityKeeper

				if tc.numPositions > 0 {
					// Initialize lower and upper ticks with empty uptime trackers
					s.initializeTick(s.Ctx, tc.positionParams.lowerTick, tc.positionParams.liquidity, cl.EmptyCoins, wrapUptimeTrackers(uptimeHelper.emptyExpectedAccumValues), true)
					s.initializeTick(s.Ctx, tc.positionParams.upperTick, tc.positionParams.liquidity, cl.EmptyCoins, wrapUptimeTrackers(uptimeHelper.emptyExpectedAccumValues), false)

					if tc.existingAccumLiquidity != nil {
						s.addLiquidityToUptimeAccumulators(s.Ctx, validPoolId, tc.existingAccumLiquidity, tc.positionParams.positionId+1)
					}

					// Initialize all positions
					for i := 0; i < tc.numPositions; i++ {
						err := clKeeper.InitOrUpdatePosition(s.Ctx, validPoolId, ownerWithValidPosition, tc.positionParams.lowerTick, tc.positionParams.upperTick, tc.positionParams.liquidity, tc.positionParams.joinTime, uint64(i+1))
						s.Require().NoError(err)
					}
					s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(tc.timeInPosition))

					// Add to uptime growth inside range
					if tc.addedUptimeGrowthInside != nil {
						s.addUptimeGrowthInsideRange(s.Ctx, validPoolId, tc.currentTick, tc.positionParams.lowerTick, tc.positionParams.upperTick, tc.addedUptimeGrowthInside)
					}

					// Add to uptime growth outside range
					if tc.addedUptimeGrowthOutside != nil {
						s.addUptimeGrowthOutsideRange(s.Ctx, validPoolId, tc.currentTick, tc.positionParams.lowerTick, tc.positionParams.upperTick, tc.addedUptimeGrowthOutside)
					}
				}

				validPool.SetCurrentTick(tc.currentTick)
				err := clKeeper.SetPool(s.Ctx, validPool)
				s.Require().NoError(err)

				// Checkpoint starting balance to compare against later
				poolBalanceBeforeCollect := s.App.BankKeeper.GetAllBalances(s.Ctx, validPool.GetAddress())
				incentivesBalanceBeforeCollect := s.App.BankKeeper.GetAllBalances(s.Ctx, validPool.GetIncentivesAddress())
				ownerBalancerBeforeCollect := s.App.BankKeeper.GetAllBalances(s.Ctx, ownerWithValidPosition)

				// System under test
				incentivesClaimedQuery, incentivesForfeitedQuery, err := clKeeper.GetClaimableIncentives(s.Ctx, DefaultPositionId)

				_ = s.App.BankKeeper.GetAllBalances(s.Ctx, validPool.GetAddress())
				incentivesBalanceAfterCollect := s.App.BankKeeper.GetAllBalances(s.Ctx, validPool.GetIncentivesAddress())
				ownerBalancerAfterCollect := s.App.BankKeeper.GetAllBalances(s.Ctx, ownerWithValidPosition)

				// Ensure balances are unchanged (since this is a query)
				s.Require().Equal(incentivesBalanceBeforeCollect, incentivesBalanceAfterCollect)
				s.Require().Equal(ownerBalancerAfterCollect, ownerBalancerBeforeCollect)

				if tc.expectedError != nil {
					s.Require().ErrorContains(err, tc.expectedError.Error())
					s.Require().Equal(tc.expectedIncentivesClaimed, incentivesClaimedQuery)
					s.Require().Equal(tc.expectedForfeitedIncentives, incentivesForfeitedQuery)
				}
				actualIncentivesClaimed, actualIncetivesForfeited, _, err := clKeeper.CollectIncentives(s.Ctx, ownerWithValidPosition, DefaultPositionId)

				// Assertions
				s.Require().Equal(incentivesClaimedQuery, actualIncentivesClaimed)
				s.Require().Equal(incentivesForfeitedQuery, actualIncetivesForfeited)

				poolBalanceAfterCollect := s.App.BankKeeper.GetAllBalances(s.Ctx, validPool.GetAddress())
				incentivesBalanceAfterCollect = s.App.BankKeeper.GetAllBalances(s.Ctx, validPool.GetIncentivesAddress())
				ownerBalancerAfterCollect = s.App.BankKeeper.GetAllBalances(s.Ctx, ownerWithValidPosition)

				// Ensure pool balances are unchanged independent of error.
				s.Require().Equal(poolBalanceBeforeCollect, poolBalanceAfterCollect)

				if tc.expectedError != nil {
					s.Require().ErrorContains(err, tc.expectedError.Error())
					s.Require().Equal(tc.expectedIncentivesClaimed, actualIncentivesClaimed)
					s.Require().Equal(tc.expectedForfeitedIncentives, actualIncetivesForfeited)

					// Ensure balances are unchanged
					s.Require().Equal(incentivesBalanceBeforeCollect, incentivesBalanceAfterCollect)
					s.Require().Equal(ownerBalancerAfterCollect, ownerBalancerBeforeCollect)
					return
				}

				// Ensure claimed amount is correct
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedIncentivesClaimed.String(), actualIncentivesClaimed.String())
				s.Require().Equal(tc.expectedForfeitedIncentives.String(), actualIncetivesForfeited.String())

				// Ensure balances are updated by the correct amounts
				// Note that we expect the forfeited incentives to remain in the pool incentives balance since they are
				// redeposited, so we only expect the diff in incentives balance to be the amount successfully claimed.
				s.Require().Equal(tc.expectedIncentivesClaimed.String(), (incentivesBalanceBeforeCollect.Sub(incentivesBalanceAfterCollect...)).String())
				s.Require().Equal(tc.expectedIncentivesClaimed.String(), (ownerBalancerAfterCollect.Sub(ownerBalancerBeforeCollect...)).String())
			})
		}
	})
}

// TestCreateIncentive tests logic around incentive creation.
// Since this function is the primary validation point for authorized uptimes, the core logic around authorized uptimes is tested here.
func (s *KeeperTestSuite) TestCreateIncentive() {
	type testCreateIncentive struct {
		poolId                   uint64
		isInvalidPoolId          bool
		useNegativeIncentiveCoin bool
		senderBalance            sdk.Coins
		recordToSet              types.IncentiveRecord
		existingRecords          []types.IncentiveRecord
		minimumGasConsumed       uint64 // default 0
		authorizedUptimes        []time.Duration

		expectedError error
	}
	tests := map[string]testCreateIncentive{
		"valid incentive record": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt(),
				),
			),
			recordToSet: incentiveRecordOne,
		},
		"record with different denom, emission rate, and min uptime": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordTwo.IncentiveRecordBody.RemainingCoin.Denom,
					incentiveRecordTwo.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt(),
				),
			),
			recordToSet:       incentiveRecordTwo,
			authorizedUptimes: types.SupportedUptimes,
		},
		"record with different start time": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt(),
				),
			),
			recordToSet: withStartTime(incentiveRecordOne, defaultStartTime.Add(time.Hour)),
		},
		"record with different incentive amount": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					osmomath.NewInt(8),
				),
			),
			recordToSet: withAmount(incentiveRecordOne, osmomath.NewDec(8)),
		},
		"existing incentive records on different uptime accumulators": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt(),
				),
			),
			recordToSet:     incentiveRecordOne,
			existingRecords: []types.IncentiveRecord{incentiveRecordTwo, incentiveRecordThree},

			// We still expect a minimum of 0 since the existing records are on other uptime accumulators
			minimumGasConsumed: uint64(0),
		},
		"existing incentive records on the same uptime accumulator": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt(),
				),
			),
			recordToSet: incentiveRecordOne,
			existingRecords: []types.IncentiveRecord{
				withMinUptime(incentiveRecordTwo, incentiveRecordOne.MinUptime),
				withMinUptime(incentiveRecordThree, incentiveRecordOne.MinUptime),
				withMinUptime(incentiveRecordFour, incentiveRecordOne.MinUptime),
			},

			// We expect # existing records * BaseGasFeeForNewIncentive. Since there are
			// three existing records on the uptime accum the new record is being added to,
			// we charge `3 * types.BaseGasFeeForNewIncentive`
			minimumGasConsumed: uint64(3 * types.BaseGasFeeForNewIncentive),
		},
		"valid incentive record on default authorized uptimes": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt(),
				),
			),
			recordToSet: incentiveRecordOne,
		},

		// Error catching
		"pool doesn't exist": {
			isInvalidPoolId: true,

			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt(),
				),
			),
			recordToSet: incentiveRecordOne,

			expectedError: types.PoolNotFoundError{PoolId: 2},
		},
		"invalid incentive coin (zero)": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					osmomath.ZeroInt(),
				),
			),
			recordToSet: withAmount(incentiveRecordOne, osmomath.ZeroDec()),

			expectedError: types.InvalidIncentiveCoinError{PoolId: 1, IncentiveCoin: sdk.NewCoin(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, osmomath.ZeroInt())},
		},
		"invalid incentive coin (negative)": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					osmomath.ZeroInt(),
				),
			),
			recordToSet:              withAmount(incentiveRecordOne, osmomath.ZeroDec()),
			useNegativeIncentiveCoin: true,

			expectedError: types.InvalidIncentiveCoinError{PoolId: 1, IncentiveCoin: sdk.Coin{Denom: incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, Amount: osmomath.NewInt(-1)}},
		},
		"start time too early": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt(),
				),
			),
			recordToSet: withStartTime(incentiveRecordOne, defaultBlockTime.Add(-1*time.Second)),

			expectedError: types.StartTimeTooEarlyError{PoolId: 1, CurrentBlockTime: defaultBlockTime, StartTime: defaultBlockTime.Add(-1 * time.Second)},
		},
		"zero emission rate": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt(),
				),
			),
			recordToSet:   withEmissionRate(incentiveRecordOne, osmomath.ZeroDec()),
			expectedError: types.NonPositiveEmissionRateError{PoolId: 1, EmissionRate: osmomath.ZeroDec()},
		},
		"negative emission rate": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt(),
				),
			),
			recordToSet: withEmissionRate(incentiveRecordOne, osmomath.NewDec(-1)),

			expectedError: types.NonPositiveEmissionRateError{PoolId: 1, EmissionRate: osmomath.NewDec(-1)},
		},
		"supported but unauthorized min uptime": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt(),
				),
			),
			recordToSet:       withMinUptime(incentiveRecordOne, time.Hour),
			authorizedUptimes: []time.Duration{time.Nanosecond, time.Minute},

			expectedError: types.InvalidMinUptimeError{PoolId: 1, MinUptime: time.Hour, AuthorizedUptimes: []time.Duration{time.Nanosecond, time.Minute}},
		},
		"unsupported min uptime": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt(),
				),
			),
			recordToSet:       withMinUptime(incentiveRecordOne, time.Hour*3),
			authorizedUptimes: types.SupportedUptimes,

			expectedError: types.InvalidMinUptimeError{PoolId: 1, MinUptime: time.Hour * 3, AuthorizedUptimes: types.SupportedUptimes},
		},
		"insufficient sender balance": {
			poolId:        defaultPoolId,
			senderBalance: sdk.NewCoins(),
			recordToSet:   incentiveRecordOne,

			expectedError: types.IncentiveInsufficientBalanceError{PoolId: 1, IncentiveDenom: incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, IncentiveAmount: incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt()},
		},
	}

	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			tc := tc
			s.Run(name, func() {
				s.SetupTest()

				// We fix blocktime to ensure tests are deterministic
				s.Ctx = s.Ctx.WithBlockTime(defaultBlockTime)

				// If specified by test case, set custom authorized uptimes (otherwise,
				// default params apply)
				if tc.authorizedUptimes != nil {
					clParams := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
					clParams.AuthorizedUptimes = tc.authorizedUptimes
					s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, clParams)
				}

				s.PrepareConcentratedPool()
				clKeeper := s.App.ConcentratedLiquidityKeeper
				s.FundAcc(s.TestAccs[0], tc.senderBalance)

				if tc.isInvalidPoolId {
					tc.poolId = tc.poolId + 1
				}

				if tc.existingRecords != nil {
					err := clKeeper.SetMultipleIncentiveRecords(s.Ctx, tc.existingRecords)
					s.Require().NoError(err)
				}

				existingGasConsumed := s.Ctx.GasMeter().GasConsumed()

				incentiveCoin := sdk.NewCoin(tc.recordToSet.IncentiveRecordBody.RemainingCoin.Denom, tc.recordToSet.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt())

				if tc.useNegativeIncentiveCoin {
					incentiveCoin = sdk.Coin{Denom: tc.recordToSet.IncentiveRecordBody.RemainingCoin.Denom, Amount: osmomath.NewInt(-1)}
				}

				// Set the next incentive record id.
				originalNextIncentiveRecordId := tc.recordToSet.IncentiveId
				clKeeper.SetNextIncentiveRecordId(s.Ctx, originalNextIncentiveRecordId)

				// system under test
				incentiveRecord, err := clKeeper.CreateIncentive(s.Ctx, tc.poolId, s.TestAccs[0], incentiveCoin, tc.recordToSet.IncentiveRecordBody.EmissionRate, tc.recordToSet.IncentiveRecordBody.StartTime, tc.recordToSet.MinUptime)

				// Assertions
				if tc.expectedError != nil {
					s.Require().ErrorContains(err, tc.expectedError.Error())

					// Ensure nothing was placed in state
					recordInState, err := clKeeper.GetIncentiveRecord(s.Ctx, tc.poolId, tc.recordToSet.MinUptime, tc.recordToSet.IncentiveId)
					s.Require().Error(err)
					s.Require().Equal(types.IncentiveRecord{}, recordInState)

					return
				}
				s.Require().NoError(err)

				// Returned incentive record should equal both to what's in state and what we expect
				recordInState, err := clKeeper.GetIncentiveRecord(s.Ctx, tc.poolId, tc.recordToSet.MinUptime, tc.recordToSet.IncentiveId)
				s.Require().NoError(err)
				s.Require().Equal(tc.recordToSet, recordInState)
				s.Require().Equal(tc.recordToSet, incentiveRecord)

				// Ensure that at least the minimum amount of gas was charged (based on number of existing incentives for current uptime)
				gasConsumed := s.Ctx.GasMeter().GasConsumed() - existingGasConsumed
				s.Require().True(gasConsumed >= tc.minimumGasConsumed)

				// Ensure that existing records aren't affected
				for _, incentiveRecord := range tc.existingRecords {
					_, err := clKeeper.GetIncentiveRecord(s.Ctx, tc.poolId, incentiveRecord.MinUptime, incentiveRecord.IncentiveId)
					s.Require().NoError(err)
				}

				// Validate that the next incentive record id was incremented
				nextIncentiveRecordId := clKeeper.GetNextIncentiveRecordId(s.Ctx)
				s.Require().Equal(originalNextIncentiveRecordId+1, nextIncentiveRecordId)
			})
		}
	})
}

// TestCreateIncentive_NewId tests that the next incentive record id is incremented
// when and a completely new incentive record is created even when the
// exact same parameters are used.
func (s *KeeperTestSuite) TestCreateIncentive_NewId() {
	s.SetupTest()

	// Initialize test parameters
	var (
		clKeeper = s.App.ConcentratedLiquidityKeeper

		pool          = s.PrepareConcentratedPool()
		poolId        = pool.GetId()
		sender        = s.TestAccs[0]
		incentiveCoin = sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(1000000000000))
		emissionRate  = osmomath.NewDecWithPrec(1, 2)
		startTime     = s.Ctx.BlockTime()
		minUptime     = types.DefaultAuthorizedUptimes[0]

		expectedIncentiveRecord = types.IncentiveRecord{
			IncentiveId: 1,
			PoolId:      poolId,
			MinUptime:   minUptime,
			IncentiveRecordBody: types.IncentiveRecordBody{
				RemainingCoin: sdk.NewDecCoinFromCoin(incentiveCoin),
				EmissionRate:  emissionRate,
				StartTime:     startTime,
			},
		}

		expectedIncentiveRecordTwo = expectedIncentiveRecord
	)
	expectedIncentiveRecordTwo.IncentiveId = 2

	// Fund the account
	s.FundAcc(sender, sdk.NewCoins(incentiveCoin.Add(incentiveCoin)))

	// Create the first incentive with given parameters
	actualIncentiveRecord, err := clKeeper.CreateIncentive(s.Ctx, poolId, sender, incentiveCoin, emissionRate, startTime, minUptime)
	s.Require().NoError(err)
	s.Require().Equal(expectedIncentiveRecord, actualIncentiveRecord)

	// Create the second incentive with the same parameters
	actualIncentiveRecord, err = clKeeper.CreateIncentive(s.Ctx, poolId, sender, incentiveCoin, emissionRate, startTime, minUptime)
	s.Require().NoError(err)
	s.Require().Equal(expectedIncentiveRecordTwo, actualIncentiveRecord)

	// Get all incentive records from state for the pool
	actualIncentiveRecords, err := clKeeper.GetAllIncentiveRecordsForPool(s.Ctx, poolId)
	s.Require().NoError(err)

	// Ensure that the incentive records are as expected
	expectedRecords := []types.IncentiveRecord{expectedIncentiveRecord, expectedIncentiveRecordTwo}
	s.Require().Equal(expectedRecords, actualIncentiveRecords)

	// Get all incentive records from state for the pool and minUptime
	actualIncentiveRecords, err = clKeeper.GetAllIncentiveRecordsForUptime(s.Ctx, poolId, minUptime)
	s.Require().NoError(err)

	// Ensure that the incentive records are as expected
	s.Require().Equal(expectedRecords, actualIncentiveRecords)
}

// TestUpdateAccumAndClaimRewards runs basic sanity checks on accumulator update and claiming logic, testing a simple happy path invariant.
// Both claiming and updating functionality is tested more thoroughly in each function's respective unit tests.
func (s *KeeperTestSuite) TestUpdateAccumAndClaimRewards() {
	validPositionKey := types.KeySpreadRewardPositionAccumulator(1)
	invalidPositionKey := types.KeySpreadRewardPositionAccumulator(2)
	tests := map[string]struct {
		poolId             uint64
		growthInside       sdk.DecCoins
		growthOutside      sdk.DecCoins
		invalidPositionKey bool
		expectError        error
	}{
		"happy path": {
			growthInside:  oneEthCoins.Add(oneEthCoins...),
			growthOutside: oneEthCoins,
		},
		"error: non existent position": {
			growthOutside:      oneEthCoins,
			invalidPositionKey: true,
			expectError:        accum.NoPositionError{Name: invalidPositionKey},
		},
	}
	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			tc := tc
			s.Run(name, func() {
				s.SetupTest()
				poolSpreadRewardsAccumulator := s.prepareSpreadRewardsAccumulator()
				positionKey := validPositionKey

				// Initialize position accumulator.
				err := poolSpreadRewardsAccumulator.NewPosition(positionKey, osmomath.OneDec(), nil)
				s.Require().NoError(err)

				// Record the initial position accumulator value.
				positionPre, err := accum.GetPosition(poolSpreadRewardsAccumulator, positionKey)
				s.Require().NoError(err)

				// If the test case requires an invalid position key, set it.
				if tc.invalidPositionKey {
					positionKey = invalidPositionKey
				}

				poolSpreadRewardsAccumulator.AddToAccumulator(tc.growthOutside.Add(tc.growthInside...))

				// System under test.
				amountClaimed, _, err := cl.UpdateAccumAndClaimRewards(poolSpreadRewardsAccumulator, positionKey, tc.growthOutside)

				if tc.expectError != nil {
					s.Require().ErrorIs(err, tc.expectError)
					return
				}
				s.Require().NoError(err)

				// We expect claimed rewards to be equal to growth inside
				expectedCoins := sdk.NormalizeCoins(tc.growthInside)
				s.Require().Equal(expectedCoins, amountClaimed)

				// Record the final position accumulator value.
				positionPost, err := accum.GetPosition(poolSpreadRewardsAccumulator, positionKey)
				s.Require().NoError(err)

				// Check that the difference between the new and old position accumulator values is equal to the growth inside (since
				// we recalibrate the position accum value after claiming).
				positionAccumDelta := positionPost.AccumValuePerShare.Sub(positionPre.AccumValuePerShare)
				s.Require().Equal(tc.growthInside, positionAccumDelta)
			})
		}
	})
}

// checkForfeitedCoinsByUptime checks that the sum of forfeited coins by uptime matches the expected total forfeited coins.
// It adds up the Coins corresponding to each uptime in the map and asserts that the result is equal to the input totalForfeitedCoins.
func (s *KeeperTestSuite) checkForfeitedCoinsByUptime(totalForfeitedCoins sdk.Coins, scaledForfeitedCoinsByUptime []sdk.Coins) {
	// Exit early if scaledForfeitedCoinsByUptime is empty
	if len(scaledForfeitedCoinsByUptime) == 0 {
		s.Require().Equal(totalForfeitedCoins, sdk.NewCoins())
		return
	}

	forfeitedCoins := sdk.NewCoins()
	// Iterate through uptime indexes and add up the forfeited coins from each
	// We unfortunately need to iterate through each coin individually to properly scale down the amount
	// (doing it in bulk leads to inconsistent rounding error)
	for uptimeIndex := range types.SupportedUptimes {
		for _, coin := range scaledForfeitedCoinsByUptime[uptimeIndex] {
			// Scale down the actual forfeited coin amount
			scaledDownAmount := cl.ScaleDownIncentiveAmount(coin.Amount, cl.PerUnitLiqScalingFactor)
			forfeitedCoins = forfeitedCoins.Add(sdk.NewCoin(coin.Denom, scaledDownAmount))
		}
	}

	s.Require().Equal(totalForfeitedCoins, forfeitedCoins, "Total forfeited coins do not match the sum of forfeited coins by uptime after scaling down")
}

// Note that the non-forfeit cases are thoroughly tested in `TestCollectIncentives`
func (s *KeeperTestSuite) TestQueryAndClaimAllIncentives() {
	uptimeHelper := getExpectedUptimes()
	defaultSender := s.TestAccs[0]
	tests := map[string]struct {
		numShares         osmomath.Dec
		positionIdCreate  uint64
		positionIdClaim   uint64
		defaultJoinTime   bool
		growthInside      []sdk.DecCoins
		growthOutside     []sdk.DecCoins
		forfeitIncentives bool
		expectedError     error
	}{
		"happy path: claim rewards without forfeiting": {
			positionIdCreate: DefaultPositionId,
			positionIdClaim:  DefaultPositionId,
			defaultJoinTime:  true,
			growthInside:     uptimeHelper.hundredTokensMultiDenom,
			growthOutside:    uptimeHelper.twoHundredTokensMultiDenom,
			numShares:        osmomath.OneDec(),
		},
		"claim and forfeit rewards (2 shares)": {
			positionIdCreate:  DefaultPositionId,
			positionIdClaim:   DefaultPositionId,
			defaultJoinTime:   true,
			growthInside:      uptimeHelper.hundredTokensMultiDenom,
			growthOutside:     uptimeHelper.twoHundredTokensMultiDenom,
			forfeitIncentives: true,
			numShares:         osmomath.NewDec(2),
		},
		"claim and forfeit rewards when no rewards have accrued": {
			positionIdCreate:  DefaultPositionId,
			positionIdClaim:   DefaultPositionId,
			defaultJoinTime:   true,
			forfeitIncentives: true,
			numShares:         osmomath.OneDec(),
		},
		"claim and forfeit rewards with varying amounts and different denoms": {
			positionIdCreate:  DefaultPositionId,
			positionIdClaim:   DefaultPositionId,
			defaultJoinTime:   true,
			growthInside:      uptimeHelper.varyingTokensMultiDenom,
			growthOutside:     uptimeHelper.varyingTokensSingleDenom,
			forfeitIncentives: true,
			numShares:         osmomath.OneDec(),
		},

		// error catching

		"error: non existent position": {
			positionIdCreate: DefaultPositionId,
			positionIdClaim:  DefaultPositionId + 1, // non existent position
			defaultJoinTime:  true,
			growthInside:     uptimeHelper.hundredTokensMultiDenom,
			growthOutside:    uptimeHelper.twoHundredTokensMultiDenom,
			numShares:        osmomath.OneDec(),

			expectedError: types.PositionIdNotFoundError{PositionId: DefaultPositionId + 1},
		},

		"error: negative duration": {
			positionIdCreate: DefaultPositionId,
			positionIdClaim:  DefaultPositionId,
			defaultJoinTime:  false,
			growthInside:     uptimeHelper.hundredTokensMultiDenom,
			growthOutside:    uptimeHelper.twoHundredTokensMultiDenom,
			numShares:        osmomath.OneDec(),

			expectedError: types.NegativeDurationError{Duration: time.Hour * 336 * -1},
		},
	}

	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			tc := tc
			s.Run(name, func() {
				s.SetupTest()
				clPool := s.PrepareConcentratedPool()
				bankKeeper := s.App.BankKeeper
				accountKeeper := s.App.AccountKeeper

				joinTime := s.Ctx.BlockTime()
				if !tc.defaultJoinTime {
					joinTime = joinTime.AddDate(0, 0, 28)
				}

				// Initialize position
				err := s.Clk.InitOrUpdatePosition(s.Ctx, validPoolId, defaultSender, DefaultLowerTick, DefaultUpperTick, tc.numShares, joinTime, tc.positionIdCreate)
				s.Require().NoError(err)

				clPool.SetCurrentTick(DefaultCurrTick)
				if tc.growthOutside != nil {
					s.addUptimeGrowthOutsideRange(s.Ctx, validPoolId, DefaultCurrTick, DefaultLowerTick, DefaultUpperTick, tc.growthOutside)
				}

				if tc.growthInside != nil {
					s.addUptimeGrowthInsideRange(s.Ctx, validPoolId, DefaultCurrTick, DefaultLowerTick, DefaultUpperTick, tc.growthInside)
				}

				err = s.Clk.SetPool(s.Ctx, clPool)
				s.Require().NoError(err)

				preCommunityPoolBalance := bankKeeper.GetAllBalances(s.Ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName))

				// Store initial pool and sender balances for comparison later
				initSenderBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, defaultSender)
				initPoolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, clPool.GetAddress())

				largestSupportedUptime := s.Clk.GetLargestSupportedUptimeDuration(s.Ctx)
				if !tc.forfeitIncentives {
					// Let enough time elapse for the position to accrue rewards for all supported uptimes.
					s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(largestSupportedUptime))
				}

				// --- System under test ---
				amountClaimedQuery, amountForfeitedQuery, err := s.Clk.GetClaimableIncentives(s.Ctx, tc.positionIdClaim)

				// Pull new balances for comparison
				newSenderBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, defaultSender)
				newPoolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, clPool.GetAddress())

				if tc.expectedError != nil {
					s.Require().ErrorIs(err, tc.expectedError)
				}

				// Ensure balances have not been mutated (since this is a query)
				s.Require().Equal(initSenderBalances, newSenderBalances)
				s.Require().Equal(initPoolBalances, newPoolBalances)

				amountClaimed, totalAmountForfeited, scaledAmountForfeitedByUptime, err := s.Clk.PrepareClaimAllIncentivesForPosition(s.Ctx, tc.positionIdClaim)

				// --- Assertions ---

				// Pull new balances for comparison
				newSenderBalances = s.App.BankKeeper.GetAllBalances(s.Ctx, defaultSender)
				newPoolBalances = s.App.BankKeeper.GetAllBalances(s.Ctx, clPool.GetAddress())

				s.Require().Equal(amountClaimedQuery, amountClaimed)
				s.Require().Equal(amountForfeitedQuery, totalAmountForfeited)
				s.checkForfeitedCoinsByUptime(totalAmountForfeited, scaledAmountForfeitedByUptime)

				if tc.expectedError != nil {
					s.Require().ErrorIs(err, tc.expectedError)

					// Ensure balances have not been mutated
					s.Require().Equal(initSenderBalances, newSenderBalances)
					s.Require().Equal(initPoolBalances, newPoolBalances)
					return
				}
				s.Require().NoError(err)

				// Prepare claim does not do any bank sends, so ensure that the community pool in unchanged
				if tc.forfeitIncentives {
					postCommunityPoolBalance := bankKeeper.GetAllBalances(s.Ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName))
					communityPoolBalanceDelta := postCommunityPoolBalance.Sub(preCommunityPoolBalance...)
					s.Require().Equal(sdk.Coins{}, amountClaimed)
					s.Require().Equal("", communityPoolBalanceDelta.String())
				} else {
					// We expect claimed rewards to be equal to growth inside
					expectedCoins := sdk.Coins(nil)
					for _, growthInside := range tc.growthInside {
						expectedCoins = expectedCoins.Add(sdk.NormalizeCoins(growthInside)...)
					}
					s.Require().Equal(expectedCoins.String(), amountClaimed.String())
					s.Require().Equal(sdk.Coins{}, totalAmountForfeited)
				}

				// Ensure balances have not been mutated
				s.Require().Equal(initSenderBalances, newSenderBalances)
				s.Require().Equal(initPoolBalances, newPoolBalances)
			})
		}
	})
}

func (s *KeeperTestSuite) TestPrepareClaimAllIncentivesForPosition() {
	testCases := []struct {
		name                              string
		blockTimeElapsed                  time.Duration
		minUptimeIncentiveRecord          time.Duration
		expectedCoins                     sdk.Coins
		expectedAccumulatorValuePostClaim sdk.DecCoins
	}{
		{
			name:                     "Claim with same blocktime",
			blockTimeElapsed:         0,
			expectedCoins:            sdk.NewCoins(),
			minUptimeIncentiveRecord: time.Nanosecond,
		},
		{
			name:                     "Claim after 1 minute, 1ns uptime",
			blockTimeElapsed:         time.Minute,
			expectedCoins:            sdk.NewCoins(sdk.NewCoin(USDC, osmomath.NewInt(59))), //  after 1min = 59.999999999901820104usdc ~ 59usdc because 1usdc emitted every second
			minUptimeIncentiveRecord: time.Nanosecond,
		},
		{
			name:                     "Claim after 1 hr, 1ns uptime",
			blockTimeElapsed:         time.Hour,
			expectedCoins:            sdk.NewCoins(sdk.NewCoin(USDC, osmomath.NewInt(3599))), //  after 1min = 59.999999999901820104usdc ~ 59usdc because 1usdc emitted every second
			minUptimeIncentiveRecord: time.Nanosecond,
		},
		{
			name:                     "Claim after 24 hours, 1ns uptime",
			blockTimeElapsed:         time.Hour * 24,
			expectedCoins:            sdk.NewCoins(sdk.NewCoin(USDC, osmomath.NewInt(9999))), //  after 24hr > 2hr46min = 9999usdc.999999999901820104 ~ 9999usdc
			minUptimeIncentiveRecord: time.Nanosecond,
		},
		{
			name:                     "Claim with same blocktime",
			blockTimeElapsed:         0,
			expectedCoins:            sdk.NewCoins(),
			minUptimeIncentiveRecord: time.Hour * 24,
		},
		{
			name:                     "Claim after 1 minute, 1d uptime",
			blockTimeElapsed:         time.Minute,
			expectedCoins:            sdk.NewCoins(),
			minUptimeIncentiveRecord: time.Hour * 24,
		},
		{
			name:                     "Claim after 1 hr, 1d uptime",
			blockTimeElapsed:         time.Hour,
			expectedCoins:            sdk.NewCoins(),
			minUptimeIncentiveRecord: time.Hour * 24,
		},
		{
			name:                     "Claim after 24 hours, 1d uptime",
			blockTimeElapsed:         time.Hour * 24,
			expectedCoins:            sdk.NewCoins(sdk.NewCoin(USDC, osmomath.NewInt(9999))), //  after 24hr > 2hr46min = 9999usdc.999999999901820104 ~ 9999usdc
			minUptimeIncentiveRecord: time.Hour * 24,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Init suite for the test.
			s.SetupTest()

			requiredBalances := sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(1_000_000)), sdk.NewCoin(USDC, osmomath.NewInt(5_000_000_000)))
			s.FundAcc(s.TestAccs[0], requiredBalances)
			s.FundAcc(s.TestAccs[1], requiredBalances)

			// Create CL pool
			pool := s.PrepareConcentratedPool()

			// Set up position
			positionOneData, err := s.Clk.CreatePosition(s.Ctx, pool.GetId(), s.TestAccs[0], requiredBalances, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
			s.Require().NoError(err)

			// Set incentives for pool to ensure accumulators work correctly
			testIncentiveRecord := types.IncentiveRecord{
				PoolId: pool.GetId(),
				IncentiveRecordBody: types.IncentiveRecordBody{
					RemainingCoin: sdk.NewDecCoinFromDec(USDC, osmomath.NewDec(10_000)), // 2hr 46m to emit all incentives
					EmissionRate:  osmomath.NewDec(1),                                   // 1 per second
					StartTime:     s.Ctx.BlockTime(),
				},
				MinUptime: tc.minUptimeIncentiveRecord,
			}
			err = s.Clk.SetMultipleIncentiveRecords(s.Ctx, []types.IncentiveRecord{testIncentiveRecord})
			s.Require().NoError(err)

			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(tc.blockTimeElapsed))

			// Update the uptime accumulators to the current block time.
			// This is done to determine the exact amount of incentives we expect to be forfeited, if any.
			err = s.Clk.UpdatePoolUptimeAccumulatorsToNow(s.Ctx, pool.GetId())
			s.Require().NoError(err)

			// Retrieve the uptime accumulators for the pool.
			uptimeAccumulatorsPreClaim, err := s.Clk.GetUptimeAccumulators(s.Ctx, pool.GetId())
			s.Require().NoError(err)

			expectedForfeitedIncentives := sdk.NewCoins()

			// If the block time elapsed is less than the minUptimeIncentiveRecord, we expect to forfeit all incentives.
			// Determine what the expected forfeited incentives per share is, including the dust since we reinvest the dust when we forfeit.
			if tc.blockTimeElapsed < tc.minUptimeIncentiveRecord {
				for _, uptimeAccum := range uptimeAccumulatorsPreClaim {
					newPositionName := string(types.KeyPositionId(positionOneData.ID))
					// Check if the accumulator contains the position.
					hasPosition := uptimeAccum.HasPosition(newPositionName)
					if hasPosition {
						position, err := accum.GetPosition(uptimeAccum, newPositionName)
						s.Require().NoError(err)

						outstandingRewards := accum.GetTotalRewards(uptimeAccum, position)

						// Scale down outstanding rewards
						for _, reward := range outstandingRewards {
							reward.Amount = reward.Amount.QuoTruncateMut(cl.PerUnitLiqScalingFactor)
						}

						collectedIncentivesForUptime, _ := outstandingRewards.TruncateDecimal()

						for _, coin := range collectedIncentivesForUptime {
							expectedForfeitedIncentives = expectedForfeitedIncentives.Add(coin)
						}
					}
				}
			}

			// System under test
			collectedInc, totalForfeitedIncentives, scaledForfeitedIncentivesByUptime, err := s.Clk.PrepareClaimAllIncentivesForPosition(s.Ctx, positionOneData.ID)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedCoins.String(), collectedInc.String())
			s.Require().Equal(expectedForfeitedIncentives.String(), totalForfeitedIncentives.String())
			s.checkForfeitedCoinsByUptime(totalForfeitedIncentives, scaledForfeitedIncentivesByUptime)

			// The difference accumulator value should have increased if we forfeited incentives by claiming.
			uptimeAccumsDiffPostClaim := sdk.NewDecCoins()
			if tc.blockTimeElapsed < tc.minUptimeIncentiveRecord {
				uptimeAccumulatorsPostClaim, err := s.Clk.GetUptimeAccumulators(s.Ctx, pool.GetId())
				s.Require().NoError(err)
				for i, acc := range uptimeAccumulatorsPostClaim {
					totalSharesAccum := acc.GetTotalShares()

					uptimeAccumsDiffPostClaim = append(uptimeAccumsDiffPostClaim, acc.GetValue().MulDec(totalSharesAccum).Sub(uptimeAccumulatorsPreClaim[i].GetValue().MulDec(totalSharesAccum))...)
				}
			}

			// We expect the incentives to be forfeited and added to the accumulator
			for i, uptimeAccumDiffPostClaim := range uptimeAccumsDiffPostClaim {
				s.Require().Equal(expectedForfeitedIncentives[i].Amount, uptimeAccumDiffPostClaim.Amount)
			}

		})
	}
}

// This functional test focuses on changing liquidity in the same range and collecting incentives
// at different times.
// This is important because the final amount of incentives claimed depends on the last time when the pool
// was updated. We use this time to calculate the amount of incentives to emit into the uptime accumulators.
func (s *KeeperTestSuite) TestFunctional_ClaimIncentives_LiquidityChange_VaryingTime() {
	s.runMultipleAuthorizedUptimes(func() {
		// Init suite for the test.
		s.SetupTest()

		const (
			testFullChargeDuration = 24 * time.Hour
		)

		var (
			defaultAddress   = s.TestAccs[0]
			defaultBlockTime = time.Unix(1, 1).UTC()
		)

		s.Ctx = s.Ctx.WithBlockTime(defaultBlockTime)
		requiredBalances := sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1))

		// Set test authorized uptime params.
		clParams := s.Clk.GetParams(s.Ctx)
		clParams.AuthorizedUptimes = []time.Duration{time.Nanosecond, testFullChargeDuration}
		s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, clParams)

		// Fund accounts twice because two positions are created.
		s.FundAcc(defaultAddress, requiredBalances.Add(requiredBalances...))

		// Create CL pool
		pool := s.PrepareConcentratedPool()

		expectedAmount := osmomath.NewInt(60 * 60 * 24) // 1 day in seconds * 1 per second

		oneUUSDCCoin := sdk.NewCoin(USDC, osmomath.OneInt())
		// -1 for acceptable rounding error
		expectedCoinsPerFullCharge := sdk.NewCoins(sdk.NewCoin(USDC, expectedAmount.Sub(osmomath.OneInt())))
		expectedHalfOfExpectedCoinsPerFullCharge := sdk.NewCoins(sdk.NewCoin(USDC, expectedAmount.QuoRaw(2).Sub(osmomath.OneInt())))

		// Multiplied by 3 because we change the block time 3 times and claim
		// 1. by directly calling CollectIncentives
		// 2. by calling WithdrawPosition
		// 3. by calling CollectIncentives
		s.FundAcc(pool.GetIncentivesAddress(), sdk.NewCoins(sdk.NewCoin(USDC, expectedAmount.Mul(osmomath.NewInt(3)))))
		// Set incentives for pool to ensure accumulators work correctly
		testIncentiveRecord := types.IncentiveRecord{
			PoolId: 1,
			IncentiveRecordBody: types.IncentiveRecordBody{
				RemainingCoin: sdk.NewDecCoinFromDec(USDC, osmomath.NewDec(1000000000000000000)),
				EmissionRate:  osmomath.NewDec(1), // 1 per second
				StartTime:     defaultBlockTime,
			},
			MinUptime: time.Nanosecond,
		}
		err := s.App.ConcentratedLiquidityKeeper.SetMultipleIncentiveRecords(s.Ctx, []types.IncentiveRecord{testIncentiveRecord})
		s.Require().NoError(err)

		// Set up position
		positionOneData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, defaultPoolId, defaultAddress, DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
		s.Require().NoError(err)

		// Increase block time by the fully charged duration (first time)
		s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(testFullChargeDuration))

		// Claim incentives.
		collected, _, _, err := s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, defaultAddress, positionOneData.ID)
		s.Require().NoError(err)
		s.Require().Equal(expectedCoinsPerFullCharge.String(), collected.String())

		// Increase block time by the fully charged duration (second time)
		s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(testFullChargeDuration))

		// Create another position
		positionTwoData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, defaultPoolId, defaultAddress, DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
		s.Require().NoError(err)

		// Increase block time by the fully charged duration (third time)
		s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(testFullChargeDuration))

		// Claim for second position. Must only claim half of the original expected amount since now there are 2 positions.
		collected, _, _, err = s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, defaultAddress, positionTwoData.ID)
		s.Require().NoError(err)
		s.Require().Equal(expectedHalfOfExpectedCoinsPerFullCharge.String(), collected.String())

		// Claim for first position and observe that claims full expected charge for the period between 1st claim and 2nd position creation
		// and half of the full charge amount since the 2nd position was created.
		collected, _, _, err = s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, defaultAddress, positionOneData.ID)
		s.Require().NoError(err)
		// Note, adding one since both expected amounts already subtract one (-2 in total)
		s.Require().Equal(expectedCoinsPerFullCharge.Add(expectedHalfOfExpectedCoinsPerFullCharge.Add(oneUUSDCCoin)...).String(), collected.String())
	})
}

// TestGetAllIncentiveRecordsForUptime tests getting all incentive records for a given uptime, regardless of whether it is authorized.
func (s *KeeperTestSuite) TestGetAllIncentiveRecordsForUptime() {
	invalidPoolId := uint64(2)
	tests := map[string]struct {
		poolIncentiveRecords []types.IncentiveRecord
		requestedUptime      time.Duration

		recordsDoNotExist bool
		poolDoesNotExist  bool

		expectedRecords []types.IncentiveRecord
		expectedError   error
	}{
		"happy path: single record on one uptime": {
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},
			requestedUptime:      incentiveRecordOne.MinUptime,

			expectedRecords: []types.IncentiveRecord{incentiveRecordOne},
		},
		"records across different uptimes": {
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo},
			requestedUptime:      incentiveRecordOne.MinUptime,

			expectedRecords: []types.IncentiveRecord{incentiveRecordOne},
		},
		"multiple records on one uptime, existing records on other uptimes": {
			poolIncentiveRecords: []types.IncentiveRecord{
				// records on requested uptime
				incentiveRecordOne, withMinUptime(incentiveRecordTwo, incentiveRecordOne.MinUptime),

				// records on other uptimes
				incentiveRecordTwo, incentiveRecordThree,
			},
			requestedUptime: incentiveRecordOne.MinUptime,

			expectedRecords: []types.IncentiveRecord{incentiveRecordOne, withMinUptime(incentiveRecordTwo, incentiveRecordOne.MinUptime)},
		},
		"no records on pool": {
			recordsDoNotExist: true,
			requestedUptime:   incentiveRecordOne.MinUptime,

			expectedRecords: []types.IncentiveRecord{},
		},
		"records on pool but none for requested uptime": {
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},
			requestedUptime:      incentiveRecordTwo.MinUptime,

			expectedRecords: []types.IncentiveRecord{},
		},

		// Error catching

		"unsupported uptime": {
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},
			requestedUptime:      time.Hour + time.Second,

			expectedRecords: []types.IncentiveRecord{},
			expectedError:   types.InvalidUptimeIndexError{MinUptime: time.Hour + time.Second, SupportedUptimes: types.SupportedUptimes},
		},
		"pool does not exist": {
			poolDoesNotExist: true,

			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},
			requestedUptime:      incentiveRecordOne.MinUptime,

			expectedRecords: []types.IncentiveRecord{},
			expectedError:   types.PoolNotFoundError{PoolId: invalidPoolId},
		},
	}
	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			tc := tc
			s.Run(name, func() {
				// --- Setup test env ---

				s.SetupTest()
				clKeeper := s.App.ConcentratedLiquidityKeeper

				// Set up pool and records unless tests requests invalid pool
				poolId := invalidPoolId
				if !tc.poolDoesNotExist {
					clPool := s.PrepareConcentratedPool()
					poolId = clPool.GetId()

					// Set incentive records across all relevant uptime accumulators
					if !tc.recordsDoNotExist {
						err := clKeeper.SetMultipleIncentiveRecords(s.Ctx, tc.poolIncentiveRecords)
						s.Require().NoError(err)
					}
				}

				// --- System under test ---

				retrievedRecords, err := clKeeper.GetAllIncentiveRecordsForUptime(s.Ctx, poolId, tc.requestedUptime)

				// --- Assertions ---

				s.Require().Equal(tc.expectedRecords, retrievedRecords)

				if tc.expectedError != nil {
					s.Require().ErrorContains(err, tc.expectedError.Error())
					return
				}
				s.Require().NoError(err)

				// --- Invariant testing ---

				// Sum of records on all supported uptime accumulators should be equal to the initially set records on the pool
				retrievedRecordsByUptime := make([]types.IncentiveRecord, len(types.SupportedUptimes))

				for _, supportedUptime := range types.SupportedUptimes {
					curUptimeRecords, err := clKeeper.GetAllIncentiveRecordsForUptime(s.Ctx, poolId, supportedUptime)
					s.Require().NoError(err)

					retrievedRecordsByUptime = append(retrievedRecordsByUptime, curUptimeRecords...)
				}
			})
		}
	})
}

// TestFindUptimeIndex tests finding the index of a given uptime in the supported uptimes constant slice.
func (s *KeeperTestSuite) TestFindUptimeIndex() {
	tests := map[string]struct {
		requestedUptime time.Duration

		expectedUptimeIndex int
		expectedError       error
	}{
		"happy path: supported uptime": {
			requestedUptime: types.SupportedUptimes[0],

			expectedUptimeIndex: 0,
		},
		"unsupported uptime": {
			requestedUptime: time.Hour + time.Second,

			expectedUptimeIndex: -1,
			expectedError:       types.InvalidUptimeIndexError{MinUptime: time.Hour + time.Second, SupportedUptimes: types.SupportedUptimes},
		},
	}
	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			tc := tc
			s.Run(name, func() {
				retrievedUptimeIndex, err := cl.FindUptimeIndex(tc.requestedUptime)

				s.Require().Equal(tc.expectedUptimeIndex, retrievedUptimeIndex)
				if tc.expectedError != nil {
					s.Require().ErrorContains(err, tc.expectedError.Error())
					return
				}
				s.Require().NoError(err)
			})
		}
	})
}

func (s *KeeperTestSuite) TestGetLargestAuthorizedAndSupportedUptimes() {
	// Note: we assume the largest supported uptime is at the end of the list.
	// While we could hardcode this, this setup is backwards compatible with changes in the list.
	longestSupportedUptime := types.SupportedUptimes[len(types.SupportedUptimes)-1]
	tests := map[string]struct {
		preSetAuthorizedParams []time.Duration
		expectedAuthorized     time.Duration
		expectedSupported      time.Duration
	}{
		"All supported uptimes authorized": {
			preSetAuthorizedParams: types.SupportedUptimes,
			expectedAuthorized:     longestSupportedUptime,
		},
		"Only 1 ns authorized": {
			preSetAuthorizedParams: []time.Duration{time.Nanosecond},
			expectedAuthorized:     time.Nanosecond,
		},
		"Unordered authorized uptimes": {
			preSetAuthorizedParams: []time.Duration{time.Hour * 24 * 7, time.Nanosecond, time.Hour * 24},
			expectedAuthorized:     time.Hour * 24 * 7,
		},
		// note cannot have test case with empty authorized uptimes due to parameter validation.
	}

	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			tc := tc
			s.Run(name, func() {
				s.SetupTest()

				clParams := s.Clk.GetParams(s.Ctx)
				clParams.AuthorizedUptimes = tc.preSetAuthorizedParams
				s.Clk.SetParams(s.Ctx, clParams)

				actualAuthorized := s.Clk.GetLargestAuthorizedUptimeDuration(s.Ctx)
				actualSupported := s.Clk.GetLargestSupportedUptimeDuration(s.Ctx)

				s.Require().Equal(tc.expectedAuthorized, actualAuthorized)
				s.Require().Equal(longestSupportedUptime, actualSupported)
			})
		}
	})
}

// 2ETH
var defaultGlobalRewardGrowth = sdk.NewDecCoins(oneEth.Add(oneEth))

func (s *KeeperTestSuite) prepareSpreadRewardsAccumulator() *accum.AccumulatorObject {
	pool := s.PrepareConcentratedPool()
	testAccumulator, err := s.Clk.GetSpreadRewardAccumulator(s.Ctx, pool.GetId())
	s.Require().NoError(err)
	return testAccumulator
}

func (s *KeeperTestSuite) TestMoveRewardsToNewPositionAndDeleteOldAcc() {
	oldPos := "old"
	newPos := "new"
	emptyCoins := sdk.DecCoins(nil)

	tests := map[string]struct {
		growthOutside            sdk.DecCoins
		expectedUnclaimedRewards sdk.DecCoins
		numShares                int64
	}{
		"empty growth outside": {
			growthOutside:            emptyCoins,
			numShares:                1,
			expectedUnclaimedRewards: defaultGlobalRewardGrowth,
		},
		"default global growth equals growth outside": {
			growthOutside:            defaultGlobalRewardGrowth,
			numShares:                1,
			expectedUnclaimedRewards: emptyCoins,
		},
		"global growth outside half of global": {
			growthOutside:            defaultGlobalRewardGrowth.QuoDec(osmomath.NewDec(2)),
			numShares:                1,
			expectedUnclaimedRewards: defaultGlobalRewardGrowth.QuoDec(osmomath.NewDec(2)),
		},
		"multiple shares, partial growth outside": {
			growthOutside:            defaultGlobalRewardGrowth.QuoDec(osmomath.NewDec(2)),
			numShares:                2,
			expectedUnclaimedRewards: defaultGlobalRewardGrowth,
		},
	}

	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			tc := tc
			s.Run(name, func() {
				s.SetupTest()

				// Get accumulator. The fact that its a fee accumulator is irrelevant for this test.
				testAccumulator := s.prepareSpreadRewardsAccumulator()

				err := testAccumulator.NewPosition(oldPos, osmomath.NewDec(tc.numShares), nil)
				s.Require().NoError(err)

				// 2 shares is chosen arbitrarily. It is not relevant for this test.
				err = testAccumulator.NewPosition(newPos, osmomath.NewDec(2), nil)
				s.Require().NoError(err)

				// Grow the global rewards.
				testAccumulator.AddToAccumulator(defaultGlobalRewardGrowth)

				err = cl.MoveRewardsToNewPositionAndDeleteOldAcc(s.Ctx, testAccumulator, oldPos, newPos, tc.growthOutside)
				s.Require().NoError(err)

				// Check the old accumulator is now deleted.
				hasPosition := testAccumulator.HasPosition(oldPos)
				s.Require().False(hasPosition)

				// Check that the new accumulator has the correct amount of rewards in unclaimed rewards.
				newPositionAccumulator, err := testAccumulator.GetPosition(newPos)
				s.Require().NoError(err)

				// Validate unclaimed rewards
				s.Require().Equal(tc.expectedUnclaimedRewards, newPositionAccumulator.UnclaimedRewardsTotal)

				// Validate that position accumulator equals to the growth inside.
				s.Require().Equal(testAccumulator.GetValue().Sub(tc.growthOutside), newPositionAccumulator.AccumValuePerShare)
			})
		}
	})
}

// This should fail because the old position does not exist.
func (s *KeeperTestSuite) TestMoveRewardsToSamePositionAndDeleteOldAcc() {
	posName := "pos name"
	expectedError := types.ModifySamePositionAccumulatorError{PositionAccName: posName}

	testAccumulator := s.prepareSpreadRewardsAccumulator()
	err := cl.MoveRewardsToNewPositionAndDeleteOldAcc(s.Ctx, testAccumulator, posName, posName, defaultGlobalRewardGrowth)
	s.Require().ErrorIs(err, expectedError)
}

// TestGetUptimeTrackerValues tests getter for tick-level uptime trackers.
func (s *KeeperTestSuite) TestGetUptimeTrackerValues() {
	foo100 := sdk.NewDecCoins(sdk.NewDecCoin("foo", osmomath.NewInt(100)))
	testCases := []struct {
		name           string
		input          []model.UptimeTracker
		expectedOutput []sdk.DecCoins
	}{
		{
			name:           "Empty uptime tracker",
			input:          []model.UptimeTracker{},
			expectedOutput: []sdk.DecCoins{},
		},
		{
			name:           "One uptime tracker",
			input:          []model.UptimeTracker{{UptimeGrowthOutside: foo100}},
			expectedOutput: []sdk.DecCoins{foo100},
		},
		{
			name: "Multiple uptime trackers",
			input: []model.UptimeTracker{
				{UptimeGrowthOutside: foo100},
				{UptimeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin("fooa", osmomath.NewInt(100)))},
				{UptimeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin("foob", osmomath.NewInt(100)))},
			},
			expectedOutput: []sdk.DecCoins{
				foo100,
				sdk.NewDecCoins(sdk.NewDecCoin("fooa", osmomath.NewInt(100))),
				sdk.NewDecCoins(sdk.NewDecCoin("foob", osmomath.NewInt(100))),
			},
		},
	}

	s.runMultipleAuthorizedUptimes(func() {
		for _, tc := range testCases {
			s.Run(tc.name, func() {
				result := cl.GetUptimeTrackerValues(tc.input)
				s.Require().Equal(tc.expectedOutput, result)
			})
		}
	})
}

func (s *KeeperTestSuite) TestGetIncentiveRecordSerialized() {
	tests := []struct {
		name                    string
		poolIdToQuery           uint64
		paginationLimit         uint64
		expectedNumberOfRecords int
	}{
		{
			name:                    "Get incentive records from a valid pool",
			poolIdToQuery:           1,
			expectedNumberOfRecords: 1,
			paginationLimit:         10,
		},
		{
			name:                    "Get many incentive records from a valid pool",
			poolIdToQuery:           1,
			expectedNumberOfRecords: 3,
			paginationLimit:         10,
		},
		{
			name:                    "Get all incentive records from an invalid pool",
			poolIdToQuery:           2,
			expectedNumberOfRecords: 0,
			paginationLimit:         10,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {

			s.SetupTest()
			k := s.App.ConcentratedLiquidityKeeper

			pool := s.PrepareConcentratedPool()

			for i := 0; i < test.expectedNumberOfRecords; i++ {
				testIncentiveRecord := types.IncentiveRecord{
					PoolId: pool.GetId(),
					IncentiveRecordBody: types.IncentiveRecordBody{
						RemainingCoin: sdk.NewDecCoinFromDec(USDC, osmomath.NewDec(1000)),
						EmissionRate:  osmomath.NewDec(1), // 1 per second
						StartTime:     defaultBlockTime,
					},
					MinUptime:   time.Nanosecond,
					IncentiveId: uint64(i + 1), // increase by 1 every iteration so that we have test.expectedNumberOfRecords unique records
				}
				err := s.App.ConcentratedLiquidityKeeper.SetIncentiveRecord(s.Ctx, testIncentiveRecord)
				s.Require().NoError(err)
			}

			paginationReq := &query.PageRequest{
				Limit:      test.paginationLimit,
				CountTotal: true,
			}

			incRecords, _, err := k.GetIncentiveRecordSerialized(s.Ctx, test.poolIdToQuery, paginationReq)
			s.Require().NoError(err)

			s.Require().Equal(test.expectedNumberOfRecords, len(incRecords))
		})
	}
}

// This test shows that there is a chance of incentives being truncated due to large liquidity value.
// We observed this in pool 1423 where both tokens have 18 decimal precision.
//
// It has been determined that no funds are at risk. The incentives are eventually distributed if either:
// a) Long time without an update to the pool state occurs (at least 51 minute with the current configuration)
// b) current tick liquidity becomes smaller
//
// As a solution, we used a scaling factor to reduce the likelihood of truncation.
// This test shows that the scaling factor is effective in reducing the likelihood of truncation.
// However, the scaling factor does not eliminate the possibility of truncation.
func (s *KeeperTestSuite) TestIncentiveTruncation() {
	s.SetupTest()

	// We multiply the current tick liqudity by this factor
	// To bring the current tick liquidity to be at the border line of truncating.
	// Then, we choose values on each side of the threshold to show that truncation is still possible
	// but, with the scaling factor, it is less likely to occur due to more room for liquidity to grow.
	currentTickLiquidityIncreaseFactor := osmomath.BigDecFromDec(cl.PerUnitLiqScalingFactor)

	// Create a pool
	pool := s.PrepareConcentratedPool()

	// 	osmosisd q concentratedliquidity incentive-records 1423 --node https://osmosis-rpc.polkachu.com:443
	// incentive_records:
	// - incentive_id: "5833"
	//   incentive_record_body:
	//     emission_rate: "9645.061724537037037037"
	//     remaining_coin:
	//       amount: "518549443.513510006462246574"
	//       denom: ibc/A8CA5EE328FA10C9519DF6057DA1F69682D28F7D0F5CCC7ECB72E3DCA2D157A4
	//     start_time: "2024-01-31T17:16:11.187417702Z"
	//   min_uptime: 0.000000001s
	//   pool_id: "1423"
	// pagination:
	//   next_key: null
	//   total: "0"
	// 24 * 60 * 60 * 9645.061724537037037037
	// 833333333.0        -<------ Initial incentives in recorrd
	incentiveCoin := sdk.NewCoin("ibc/A8CA5EE328FA10C9519DF6057DA1F69682D28F7D0F5CCC7ECB72E3DCA2D157A4", osmomath.NewInt(833333333))

	// Create a pool state simulating pool 1423. The only difference is that we force the pool state given 1 position as
	// opposed to many.
	// Liquidity around height 13,607,920 in pool 1423
	// We multiply by the scaling factor value as a sanity check that no truncation occurs within our 10^27 scaling factor choice.
	desiredLiquidity := osmomath.MustNewBigDecFromStr("180566277759640622277799341.480727726620927100").MulMut(currentTickLiquidityIncreaseFactor)
	desiredCurrentTick := int64(596)
	desiredCurrentSqrtPrice, err := math.TickToSqrtPrice(desiredCurrentTick)
	s.Require().NoError(err)

	amount0 := math.CalcAmount0Delta(desiredLiquidity.Dec(), desiredCurrentSqrtPrice, types.MaxSqrtPriceBigDec, true).Dec().TruncateInt()
	amount1 := math.CalcAmount1Delta(desiredLiquidity.Dec(), types.MinSqrtPriceBigDec, desiredCurrentSqrtPrice, true).Dec().TruncateInt()

	lpCoins := sdk.NewCoins(sdk.NewCoin(ETH, amount0), sdk.NewCoin(USDC, amount1))
	s.FundAcc(s.TestAccs[0], lpCoins)

	// LP
	positionData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, pool.GetId(), s.TestAccs[0], lpCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), types.MinInitializedTick, types.MaxTick)
	s.Require().NoError(err)

	fmt.Println("initial liquidity", positionData.Liquidity)

	// Fund the account with the incentive coin
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(incentiveCoin))

	// Set incentives for pool to ensure accumulators work correctly
	_, err = s.App.ConcentratedLiquidityKeeper.CreateIncentive(s.Ctx, pool.GetId(), s.TestAccs[0], incentiveCoin, osmomath.MustNewDecFromStr("9645.061724537037037037"), s.Ctx.BlockTime(), time.Nanosecond)
	s.Require().NoError(err)

	// The check below shows that the incentive is not claimed due to truncation
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Minute * 50))
	incentives, _, _, err := s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, s.TestAccs[0], positionData.ID)
	s.Require().NoError(err)
	s.Require().True(incentives.IsZero())

	// Incentives should now be claimed due to lack of truncation
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 6))
	incentives, _, _, err = s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, s.TestAccs[0], positionData.ID)
	s.Require().NoError(err)
	s.Require().False(incentives.IsZero())
}

// This test shows that the scaling factor is applied correctly to the total incentive amount.
// If overflow occurs, the function returns error as opposed to panicking.
func (s *KeeperTestSuite) TestScaledUpTotalIncentiveAmount() {
	scaledIncentiveAmount, err := cl.ScaleUpTotalEmittedAmount(osmomath.NewDec(1), cl.PerUnitLiqScalingFactor)
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewDec(1).Mul(cl.PerUnitLiqScalingFactor), scaledIncentiveAmount)

	_, err = cl.ScaleUpTotalEmittedAmount(oneE60Dec, cl.PerUnitLiqScalingFactor)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "overflow")
}

// This test shows that it is possible to compute the total incentives to emit without overflow.
// An error is returned if overflow occurs but no panic.
func (s *KeeperTestSuite) TestComputeTotalIncentivesToEmit() {

	oneHundredYearsSecs := osmomath.NewDec(int64((time.Hour * 24 * 365 * 100).Seconds()))

	totalIncentiveAmount, err := cl.ComputeTotalIncentivesToEmit(oneHundredYearsSecs, osmomath.NewDec(1))
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewDec(1).Mul(oneHundredYearsSecs), totalIncentiveAmount)

	// The value of 1_000_000_000_000 is hand-picked to be close to the max of 2^256 so that
	// when multiplied by 100 years, it overflows.
	_, err = cl.ComputeTotalIncentivesToEmit(oneHundredYearsSecs, oneE60Dec.MulInt64(1_000_000_000_000))
	s.Require().Error(err)
	s.Require().ErrorContains(err, "overflow")
}

// This test shows that the scaling factor is applied correctly based on the pool ID.
// If the pool ID is below or at the migration threshold, the scaling factor is 1.
// If the pool ID is above the migration threshold, the scaling factor is the per unit liquidity scaling factor.
// If he pool ID is in the overwrite map, the scaling factor is the per unit liquidity scaling factor.
func (s *KeeperTestSuite) TestGetIncentiveScalingFactorForPool() {
	// Grab an example of the overwrite pool from map
	s.Require().NotZero(len(types.MigratedIncentiveAccumulatorPoolIDs))
	s.Require().NotZero(len(types.MigratedIncentiveAccumulatorPoolIDsV24))

	var exampleOverwritePoolID uint64
	for poolID := range types.MigratedIncentiveAccumulatorPoolIDs {
		exampleOverwritePoolID = poolID
		break
	}

	var exampleOverwritePoolIDv24 uint64
	for poolIDv24 := range types.MigratedIncentiveAccumulatorPoolIDsV24 {
		exampleOverwritePoolIDv24 = poolIDv24
		break
	}

	var (
		oneDec = osmomath.OneDec()

		// Make migration threshold 1000 pool IDs higher
		migrationThreshold = exampleOverwritePoolID + 1000
	)

	s.SetupTest()

	s.App.ConcentratedLiquidityKeeper.SetIncentivePoolIDMigrationThreshold(s.Ctx, migrationThreshold)

	// One below the threshold has scaling factor of 1 (non-migrated)
	scalingFactor, err := s.App.ConcentratedLiquidityKeeper.GetIncentiveScalingFactorForPool(s.Ctx, migrationThreshold-1)
	s.Require().NoError(err)
	s.Require().Equal(oneDec, scalingFactor)

	// At threshold, scaling factor is 1 (non-migrated)
	scalingFactor, err = s.App.ConcentratedLiquidityKeeper.GetIncentiveScalingFactorForPool(s.Ctx, migrationThreshold)
	s.Require().NoError(err)
	s.Require().Equal(oneDec, scalingFactor)

	// One above the threshold has non-1 scaling factor (migrated)
	scalingFactor, err = s.App.ConcentratedLiquidityKeeper.GetIncentiveScalingFactorForPool(s.Ctx, migrationThreshold+1)
	s.Require().NoError(err)
	s.Require().Equal(cl.PerUnitLiqScalingFactor, scalingFactor)

	// Overwrite pool ID has non-1 scaling factor (migrated)
	scalingFactor, err = s.App.ConcentratedLiquidityKeeper.GetIncentiveScalingFactorForPool(s.Ctx, exampleOverwritePoolID)
	s.Require().NoError(err)
	s.Require().Equal(cl.PerUnitLiqScalingFactor, scalingFactor)

	// Overwrite pool IDv24 has non-1 scaling factor (migrated)
	scalingFactor, err = s.App.ConcentratedLiquidityKeeper.GetIncentiveScalingFactorForPool(s.Ctx, exampleOverwritePoolIDv24)
	s.Require().NoError(err)
	s.Require().Equal(cl.PerUnitLiqScalingFactor, scalingFactor)
}

// scaleUptimeAccumulators scales the uptime accumulators by the scaling factor.
// This is to avoid truncation to zero in core logic when the liquidity is large.
func (s *KeeperTestSuite) scaleUptimeAccumulators(uptimeAccumulatorsToScale []sdk.DecCoins) []sdk.DecCoins {
	growthCopy := make([]sdk.DecCoins, len(uptimeAccumulatorsToScale))
	for i, growth := range uptimeAccumulatorsToScale {
		growthCopy[i] = make(sdk.DecCoins, len(growth))
		for j, coin := range growth {
			growthCopy[i][j].Denom = coin.Denom
			growthCopy[i][j].Amount = coin.Amount.MulTruncate(cl.PerUnitLiqScalingFactor)
		}
	}

	return growthCopy
}

// assertUptimeAccumsEmpty asserts that the uptime accumulators for the given pool are empty.
func (s *KeeperTestSuite) assertUptimeAccumsEmpty(poolId uint64) {
	uptimeAccums, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, poolId)
	s.Require().NoError(err)

	// Ensure uptime accums remain empty
	for _, accum := range uptimeAccums {
		s.Require().Equal(sdk.NewDecCoins(), sdk.NewDecCoins(accum.GetValue()...))
	}
}

// TestRedepositForfeitedIncentives tests the redeposit of forfeited incentives into uptime accumulators.
// In the cases where the pool has active liquidity and the incentives need to be redeposited,
// it asserts that forfeitedIncentives / activeLiquidity was deposited into the accumulators.
// In the cases where the pool has no active liquidity, it asserts that the forfeited incentives were sent to the owner.
func (s *KeeperTestSuite) TestRedepositForfeitedIncentives() {
	tests := map[string]struct {
		setupPoolWithActiveLiquidity bool
		forfeitedIncentives          []sdk.Coins
		expectedError                error
	}{
		"No forfeited incentives": {
			setupPoolWithActiveLiquidity: true,
			forfeitedIncentives:          []sdk.Coins{sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins()},
		},
		"With active liquidity - forfeited incentives redeposited": {
			setupPoolWithActiveLiquidity: true,
			forfeitedIncentives:          []sdk.Coins{{sdk.NewCoin("foo", osmomath.NewInt(12345))}, sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins()},
		},
		"Multiple forfeited incentives redeposited": {
			setupPoolWithActiveLiquidity: true,
			forfeitedIncentives:          []sdk.Coins{sdk.NewCoins(), {sdk.NewCoin("bar", osmomath.NewInt(54321))}, sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), {sdk.NewCoin("foo", osmomath.NewInt(12345))}},
		},
		"All slots filled with forfeited incentives": {
			setupPoolWithActiveLiquidity: true,
			forfeitedIncentives:          []sdk.Coins{{sdk.NewCoin("foo", osmomath.NewInt(10000))}, {sdk.NewCoin("bar", osmomath.NewInt(20000))}, {sdk.NewCoin("baz", osmomath.NewInt(30000))}, {sdk.NewCoin("qux", osmomath.NewInt(40000))}, {sdk.NewCoin("quux", osmomath.NewInt(50000))}, {sdk.NewCoin("corge", osmomath.NewInt(60000))}},
		},
		"No active liquidity with no forfeited incentives": {
			setupPoolWithActiveLiquidity: false,
			forfeitedIncentives:          []sdk.Coins{sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins()},
		},
		"No active liquidity with forfeited incentives sent to owner": {
			setupPoolWithActiveLiquidity: false,
			forfeitedIncentives:          []sdk.Coins{{sdk.NewCoin("foo", osmomath.NewInt(10000))}, sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins()},
		},
		"Incorrect forfeited incentives length": {
			setupPoolWithActiveLiquidity: true,
			forfeitedIncentives:          []sdk.Coins{sdk.NewCoins()}, // Incorrect length, should be len(types.SupportedUptimes)
			expectedError:                types.InvalidForfeitedIncentivesLengthError{ForfeitedIncentivesLength: 1, ExpectedLength: len(types.SupportedUptimes)},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			// Setup pool
			pool := s.PrepareConcentratedPool()
			poolId := pool.GetId()
			owner := s.TestAccs[0]

			if tc.setupPoolWithActiveLiquidity {
				// Add position to ensure pool has active liquidity
				s.SetupDefaultPosition(poolId)
			}

			// Fund pool with forfeited incentives and track total
			totalForfeitedIncentives := sdk.NewCoins()
			for _, coins := range tc.forfeitedIncentives {
				s.FundAcc(pool.GetIncentivesAddress(), coins)
				totalForfeitedIncentives = totalForfeitedIncentives.Add(coins...)
			}

			// Get balances before the operation to compare after
			balancesBefore := s.App.BankKeeper.GetAllBalances(s.Ctx, owner)

			// --- System under test ---

			err := s.App.ConcentratedLiquidityKeeper.RedepositForfeitedIncentives(s.Ctx, poolId, owner, tc.forfeitedIncentives, totalForfeitedIncentives)

			// --- Assertions ---

			balancesAfter := s.App.BankKeeper.GetAllBalances(s.Ctx, owner)
			balanceChange := balancesAfter.Sub(balancesBefore...)

			// If an error is expected, check if it matches the expected error
			if tc.expectedError != nil {
				s.Require().ErrorContains(err, tc.expectedError.Error())

				// Check if the owner's balance did not change
				s.Require().Equal(sdk.NewCoins(), sdk.NewCoins(balanceChange...))

				// Assert that uptime accumulators remain empty
				s.assertUptimeAccumsEmpty(poolId)
				return
			}

			s.Require().NoError(err)

			// If there is no active liquidity, the forfeited incentives should be sent to the owner
			if !tc.setupPoolWithActiveLiquidity {
				// Check if the owner received the forfeited incentives
				s.Require().Equal(totalForfeitedIncentives, balanceChange)

				// Assert that uptime accumulators remain empty
				s.assertUptimeAccumsEmpty(poolId)
				return
			}

			// If there is active liquidity, the forfeited incentives should not
			// be sent to the owner, but instead redeposited into the uptime accumulators.
			s.Require().Equal(sdk.NewCoins(), sdk.NewCoins(balanceChange...))

			// Refetch updated pool and accumulators
			pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
			s.Require().NoError(err)
			uptimeAccums, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, poolId)
			s.Require().NoError(err)

			// Check if the forfeited incentives were redeposited into the uptime accumulators
			for i, accum := range uptimeAccums {
				// Check that each accumulator has the correct value of scaledForfeitedIncentives / activeLiquidity
				// Note that the function assumed the input slice is already scaled to avoid unnecessary recomputation.
				for _, forfeitedCoin := range tc.forfeitedIncentives[i] {
					expectedAmount := forfeitedCoin.Amount.ToLegacyDec().QuoTruncate(pool.GetLiquidity())
					accumAmount := accum.GetValue().AmountOf(forfeitedCoin.Denom)
					s.Require().Equal(expectedAmount, accumAmount, "Forfeited incentive amount mismatch in uptime accumulator")
				}
			}
		})
	}
}
