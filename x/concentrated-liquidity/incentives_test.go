package concentrated_liquidity_test

import (
	"errors"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	cl "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v19/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v19/x/gamm/types"
	gammmigration "github.com/osmosis-labs/osmosis/v19/x/gamm/types/migration"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v19/x/pool-incentives/types"
)

var (
	defaultPoolId     = uint64(1)
	defaultMultiplier = sdk.OneInt()

	testAddressOne   = sdk.MustAccAddressFromBech32("osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks")
	testAddressTwo   = sdk.MustAccAddressFromBech32("osmo18s5lynnmx37hq4wlrw9gdn68sg2uxp5rgk26vv")
	testAddressThree = sdk.MustAccAddressFromBech32("osmo1qwexv7c6sm95lwhzn9027vyu2ccneaqad4w8ka")
	testAddressFour  = sdk.MustAccAddressFromBech32("osmo14hcxlnwlqtq75ttaxf674vk6mafspg8xwgnn53")

	// Note: lexicographic order is denomFour, denomOne, denomThree, denomTwo
	testDenomOne   = "denomOne"
	testDenomTwo   = "denomTwo"
	testDenomThree = "denomThree"
	testDenomFour  = "denomFour"

	defaultIncentiveAmount   = sdk.NewDec(2 << 60)
	defaultIncentiveRecordId = uint64(1)

	testEmissionOne   = sdk.MustNewDecFromStr("0.000001")
	testEmissionTwo   = sdk.MustNewDecFromStr("0.0783")
	testEmissionThree = sdk.MustNewDecFromStr("165.4")
	testEmissionFour  = sdk.MustNewDecFromStr("57.93")

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
			RemainingCoin: sdk.NewDecCoinFromDec("emptyDenom", sdk.ZeroDec()),
			EmissionRate:  testEmissionFour,
			StartTime:     defaultStartTime,
		},
		MinUptime:   testUptimeFour,
		IncentiveId: defaultIncentiveRecordId + 4,
	}

	testQualifyingDepositsOne = sdk.NewInt(50)

	defaultBalancerAssets = []balancer.PoolAsset{
		{Weight: sdk.NewInt(1), Token: sdk.NewCoin("foo", sdk.NewInt(100))},
		{Weight: sdk.NewInt(1), Token: sdk.NewCoin("bar", sdk.NewInt(100))},
	}
	defaultConcentratedAssets = sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(100)), sdk.NewCoin("bar", sdk.NewInt(100)))
	defaultBalancerPoolParams = balancer.PoolParams{SwapFee: sdk.NewDec(0), ExitFee: sdk.NewDec(0)}
	invalidPoolId             = uint64(10)
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
		expUptimes.varyingTokensSingleDenom = append(expUptimes.varyingTokensSingleDenom, sdk.NewDecCoins(cl.HundredFooCoins.Add(sdk.NewDecCoin("foo", sdk.NewInt(int64(i))))))
		expUptimes.varyingTokensMultiDenom = append(expUptimes.varyingTokensMultiDenom, sdk.NewDecCoins(cl.HundredFooCoins.Add(sdk.NewDecCoin("foo", sdk.NewInt(int64(i)))), cl.HundredBarCoins.Add(sdk.NewDecCoin("bar", sdk.NewInt(int64(i*3))))))
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
func expectedIncentivesFromRate(denom string, rate sdk.Dec, timeElapsed time.Duration, qualifyingLiquidity sdk.Dec) sdk.DecCoin {
	timeInSec := sdk.NewDec(int64(timeElapsed)).Quo(sdk.MustNewDecFromStr("1000000000"))
	amount := rate.Mul(timeInSec).QuoTruncate(qualifyingLiquidity)

	return sdk.NewDecCoinFromDec(denom, amount)
}

// expectedIncentivesFromUptimeGrowth calculates the amount of incentives we expect to accrue based on uptime accumulator growth.
//
// Assumes `uptimeGrowths` represents the growths for all global uptime accums and only counts growth that `timeInPool` qualifies for
// towards result. Takes in a multiplier parameter for further versatility in testing.
//
// Returns value as truncated sdk.Coins as the primary use of this helper is testing higher level incentives functions such as claiming.
func expectedIncentivesFromUptimeGrowth(uptimeGrowths []sdk.DecCoins, positionShares sdk.Dec, timeInPool time.Duration, multiplier sdk.Int) sdk.Coins {
	// Sum up rewards from all inputs
	totalRewards := sdk.Coins(nil)
	for uptimeIndex, uptimeGrowth := range uptimeGrowths {
		if timeInPool >= types.SupportedUptimes[uptimeIndex] {
			curRewards := uptimeGrowth.MulDecTruncate(positionShares).MulDecTruncate(multiplier.ToDec())
			totalRewards = totalRewards.Add(sdk.NormalizeCoins(curRewards)...)
		}
	}

	return totalRewards
}

// chargeIncentiveRecord updates the remaining amount of the passed in incentive record to what it would be after `timeElapsed` of emissions.
func chargeIncentiveRecord(incentiveRecord types.IncentiveRecord, timeElapsed time.Duration) types.IncentiveRecord {
	secToNanoSec := int64(1000000000)
	incentivesEmitted := incentiveRecord.IncentiveRecordBody.EmissionRate.Mul(sdk.NewDec(int64(timeElapsed)).Quo(sdk.NewDec(secToNanoSec)))
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
		uptimeAccum.AddToAccumulator(addValues[uptimeIndex])
	}

	return nil
}

func withDenom(record types.IncentiveRecord, denom string) types.IncentiveRecord {
	record.IncentiveRecordBody.RemainingCoin.Denom = denom

	return record
}

func withAmount(record types.IncentiveRecord, amount sdk.Dec) types.IncentiveRecord {
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

func withEmissionRate(record types.IncentiveRecord, emissionRate sdk.Dec) types.IncentiveRecord {
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
	records, err := s.clk.GetAllIncentiveRecordsForPool(s.Ctx, poolId)
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
		qualifyingLiquidity  sdk.Dec
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
			qualifyingLiquidity:  sdk.NewDec(100),
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},

			expectedResult: sdk.DecCoins{
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{chargeIncentiveRecord(incentiveRecordOne, time.Hour)},
			expectedPass:             true,
		},
		"one incentive record, one qualifying for incentives, start time after current block time": {
			poolId:               defaultPoolId,
			accumUptime:          types.SupportedUptimes[0],
			qualifyingLiquidity:  sdk.NewDec(100),
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOneWithStartTimeAfterBlockTime},

			expectedResult:           sdk.DecCoins{},
			expectedIncentiveRecords: []types.IncentiveRecord{incentiveRecordOneWithStartTimeAfterBlockTime},
			expectedPass:             true,
		},
		"two incentive records, one with qualifying liquidity for incentives": {
			poolId:               defaultPoolId,
			accumUptime:          types.SupportedUptimes[0],
			qualifyingLiquidity:  sdk.NewDec(100),
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo},

			expectedResult: sdk.DecCoins{
				// We only expect the first incentive record to qualify
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
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
			qualifyingLiquidity: sdk.NewDec(123),

			// Time elapsed is strictly greater than the time needed to emit all incentives
			timeElapsed: time.Duration((1 << 63) - 1),
			poolIncentiveRecords: []types.IncentiveRecord{
				// We set the emission rate high enough to drain the record in one timestep
				withEmissionRate(incentiveRecordOne, sdk.NewDec(2<<60)),
			},
			recordsCleared: true,

			// We expect the fully incentive amount to be emitted
			expectedResult: sdk.DecCoins{
				sdk.NewDecCoinFromDec(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.QuoTruncate(sdk.NewDec(123))),
			},

			// Incentive record should have zero remaining amount
			expectedIncentiveRecords: []types.IncentiveRecord{withAmount(withEmissionRate(incentiveRecordOne, sdk.NewDec(2<<60)), sdk.ZeroDec())},
			expectedPass:             true,
		},

		// error catching
		"zero qualifying liquidity": {
			poolId:               defaultPoolId,
			accumUptime:          types.SupportedUptimes[0],
			qualifyingLiquidity:  sdk.NewDec(0),
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},

			expectedResult:           sdk.DecCoins{},
			expectedIncentiveRecords: []types.IncentiveRecord{},
			expectedPass:             false,
		},
		"zero time elapsed": {
			poolId:               defaultPoolId,
			accumUptime:          types.SupportedUptimes[0],
			qualifyingLiquidity:  sdk.NewDec(100),
			timeElapsed:          time.Duration(0),
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},

			expectedResult:           sdk.DecCoins{},
			expectedIncentiveRecords: []types.IncentiveRecord{},
			expectedPass:             false,
		},
		"two incentive records with same denom, different start time": {
			poolId:              defaultPoolId,
			accumUptime:         types.SupportedUptimes[0],
			qualifyingLiquidity: sdk.NewDec(100),
			timeElapsed:         time.Hour,

			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordOneWithDifferentStartTime},

			expectedResult: sdk.NewDecCoins(
				// We expect both incentive records to qualify
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate.Add(incentiveRecordOneWithDifferentStartTime.IncentiveRecordBody.EmissionRate), time.Hour, sdk.NewDec(100)), // since we have 2 records with same denom, the rate of emission went up x2
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
			qualifyingLiquidity: sdk.NewDec(100),
			timeElapsed:         time.Hour,

			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOneWithDifferentStartTime, incentiveRecordOneWithDifferentDenom},

			expectedResult: sdk.DecCoins{
				// We expect both incentive record to qualify
				expectedIncentivesFromRate(incentiveRecordOneWithDifferentStartTime.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
				expectedIncentivesFromRate(incentiveRecordOneWithDifferentDenom.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
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
			qualifyingLiquidity: sdk.NewDec(100),
			timeElapsed:         time.Hour,

			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordOneWithDifferentMinUpTime},

			expectedResult: sdk.DecCoins{
				// We expect first incentive record to qualify
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
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
			qualifyingLiquidity: sdk.NewDec(100),
			timeElapsed:         time.Hour,

			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordOneWithDifferentDenom},

			expectedResult: sdk.DecCoins{
				// We expect both incentive record to qualify
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
				expectedIncentivesFromRate(incentiveRecordOneWithDifferentDenom.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
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
			qualifyingLiquidity: sdk.NewDec(100),
			timeElapsed:         time.Hour,

			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordOneWithDifferentStartTime, incentiveRecordOneWithDifferentDenom, incentiveRecordOneWithDifferentMinUpTime},

			expectedResult: sdk.NewDecCoins(
				// We expect three incentive record to qualify for incentive
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate.Add(incentiveRecordOneWithDifferentStartTime.IncentiveRecordBody.EmissionRate), time.Hour, sdk.NewDec(100)),
				expectedIncentivesFromRate(incentiveRecordOneWithDifferentDenom.IncentiveRecordBody.RemainingCoin.Denom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
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
				actualResult, updatedPoolRecords, err := cl.CalcAccruedIncentivesForAccum(s.Ctx, tc.accumUptime, tc.qualifyingLiquidity, sdk.NewDec(int64(tc.timeElapsed)).Quo(sdk.MustNewDecFromStr("1000000000")), tc.poolIncentiveRecords)
				if tc.expectedPass {
					s.Require().NoError(err)

					s.Require().Equal(tc.expectedResult, actualResult)
					s.Require().Equal(tc.expectedIncentiveRecords, updatedPoolRecords)

					// If incentives are fully emitted, we ensure they are cleared from state
					if tc.recordsCleared {
						err := s.clk.SetMultipleIncentiveRecords(s.Ctx, updatedPoolRecords)
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

func (s *KeeperTestSuite) setupBalancerPoolWithFractionLocked(pa []balancer.PoolAsset, fraction sdk.Dec) uint64 {
	balancerPoolId := s.PrepareCustomBalancerPool(pa, defaultBalancerPoolParams)
	longestLockableDuration, err := s.App.PoolIncentivesKeeper.GetLongestLockableDuration(s.Ctx)
	s.Require().NoError(err)
	lockAmt := gammtypes.InitPoolSharesSupply.ToDec().Mul(fraction).TruncateInt()
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
	longestLockableDuration, err := s.App.PoolIncentivesKeeper.GetLongestLockableDuration(s.Ctx)
	s.Require().NoError(err)
	type updateAccumToNow struct {
		poolId                      uint64
		timeElapsed                 time.Duration
		poolIncentiveRecords        []types.IncentiveRecord
		canonicalBalancerPoolAssets []balancer.PoolAsset

		isInvalidBalancerPool    bool
		expectedIncentiveRecords []types.IncentiveRecord
		expectedError            error
	}

	validateResult := func(ctx sdk.Context, err error, tc updateAccumToNow, balancerPoolId, poolId uint64, initUptimeAccumValues []sdk.DecCoins, qualifyingBalancerLiquidity sdk.Dec, qualifyingLiquidity sdk.Dec) []sdk.DecCoins {
		if tc.expectedError != nil {
			s.Require().ErrorContains(err, tc.expectedError.Error())

			// Ensure accumulators remain unchanged
			newUptimeAccumValues, err := s.clk.GetUptimeAccumulatorValues(ctx, poolId)
			s.Require().NoError(err)
			s.Require().Equal(initUptimeAccumValues, newUptimeAccumValues)

			// Ensure incentive records remain unchanged
			updatedIncentiveRecords := s.getAllIncentiveRecordsForPool(poolId)
			s.Require().Equal(tc.poolIncentiveRecords, updatedIncentiveRecords)

			return nil
		}

		s.Require().NoError(err)

		// Get updated pool for testing purposes
		clPool, err := s.clk.GetPoolById(ctx, tc.poolId)
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
		newUptimeAccumValues, err := s.clk.GetUptimeAccumulatorValues(ctx, tc.poolId)
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
		updatedIncentiveRecords, err := s.clk.GetAllIncentiveRecordsForPool(ctx, tc.poolId)
		s.Require().NoError(err)
		s.Require().Equal(tc.expectedIncentiveRecords, updatedIncentiveRecords)

		// If applicable, get gauge for canonical balancer pool and ensure it increased by the appropriate amount.
		if tc.canonicalBalancerPoolAssets != nil {
			gaugeId, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(ctx, balancerPoolId, longestLockableDuration)
			s.Require().NoError(err)

			gauge, err := s.App.IncentivesKeeper.GetGaugeByID(ctx, gaugeId)
			s.Require().NoError(err)

			// Since balancer shares are added prior to actual emissions to the pool, they are already factored into the
			// accumulator values ("totalUptimeDeltas"). We leverage this to find the expected amount of incentives emitted
			// to the gauge.
			expectedGaugeShares := sdk.NewCoins(sdk.NormalizeCoins(totalUptimeDeltas.MulDec(qualifyingBalancerLiquidity))...)
			s.Require().Equal(expectedGaugeShares, gauge.Coins)
		}

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
		"[reward splitting] one incentive record with qualifying liquidity": {
			poolId:               defaultPoolId,
			timeElapsed:          time.Hour * 1000,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},
			canonicalBalancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("eth", sdk.NewInt(100000000))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("usdc", sdk.NewInt(100000000))},
			},

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We deduct incentives from the record for the period it emitted incentives
				chargeIncentiveRecord(incentiveRecordOne, defaultTestUptime+(time.Hour*1000)),
			},
		},
		"[reward splitting] four incentive records, only three with qualifying liquidity": {
			poolId:               defaultPoolId,
			timeElapsed:          time.Hour * 1000,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo, incentiveRecordThree, incentiveRecordFour},
			canonicalBalancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("eth", sdk.NewInt(100000000))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("usdc", sdk.NewInt(100000000))},
			},

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We only deduct from the first three incentive records since the last doesn't emit anything
				// Note that records are in ascending order by uptime index
				chargeIncentiveRecord(incentiveRecordOne, defaultTestUptime+(time.Hour*1000)),
				chargeIncentiveRecord(incentiveRecordTwo, defaultTestUptime+(time.Hour*1000)),
				chargeIncentiveRecord(incentiveRecordThree, defaultTestUptime+(time.Hour*1000)),
				// We charge even for uptimes the position has technically not qualified for since its liquidity is on
				// the accumulator.
				chargeIncentiveRecord(incentiveRecordFour, defaultTestUptime+(time.Hour*1000)),
			},
		},
		"[reward splitting] three incentive records, each with qualifying liquidity and small amount of time elapsed": {
			poolId:               defaultPoolId,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo, incentiveRecordThree},
			canonicalBalancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("eth", sdk.NewInt(432218877))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("usdc", sdk.NewInt(19836275))},
			},

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We deduct incentives from each record since there are positions for all three
				// Note that records are in ascending order by uptime index
				chargeIncentiveRecord(incentiveRecordOne, defaultTestUptime+time.Hour),
				chargeIncentiveRecord(incentiveRecordTwo, defaultTestUptime+time.Hour),
				chargeIncentiveRecord(incentiveRecordThree, defaultTestUptime+time.Hour),
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
		"invalid canonical balancer pool (incorrect denoms)": {
			poolId:                      defaultPoolId,
			timeElapsed:                 time.Hour,
			poolIncentiveRecords:        []types.IncentiveRecord{incentiveRecordOne},
			canonicalBalancerPoolAssets: defaultBalancerAssets,
			isInvalidBalancerPool:       true,

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We deduct incentives from the record for the period it emitted incentives
				chargeIncentiveRecord(incentiveRecordOne, defaultTestUptime+time.Hour),
			},
			expectedError: types.ErrInvalidBalancerPoolLiquidityError{
				ClPoolId:              1,
				BalancerPoolId:        2,
				BalancerPoolLiquidity: sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(100)), sdk.NewCoin("foo", sdk.NewInt(100))),
			},
		},
		"invalid canonical balancer pool (incorrect number of assets)": {
			poolId:               defaultPoolId,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},
			canonicalBalancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("foo", sdk.NewInt(100))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("bar", sdk.NewInt(100))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("baz", sdk.NewInt(100))},
			},
			isInvalidBalancerPool: true,

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We deduct incentives from the record for the period it emitted incentives
				chargeIncentiveRecord(incentiveRecordOne, defaultTestUptime+time.Hour),
			},
			expectedError: types.ErrInvalidBalancerPoolLiquidityError{
				ClPoolId:              1,
				BalancerPoolId:        2,
				BalancerPoolLiquidity: sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(100)), sdk.NewCoin("baz", sdk.NewInt(100)), sdk.NewCoin("foo", sdk.NewInt(100))),
			},
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

				// If applicable, create and link a canonical balancer pool
				balancerPoolId := uint64(0)
				if tc.canonicalBalancerPoolAssets != nil {
					// Create balancer pool and bond its shares
					balancerPoolId = s.setupBalancerPoolWithFractionLocked(tc.canonicalBalancerPoolAssets, sdk.OneDec())
					s.App.GAMMKeeper.OverwriteMigrationRecordsAndRedirectDistrRecords(s.Ctx,
						gammmigration.MigrationRecords{
							BalancerToConcentratedPoolLinks: []gammmigration.BalancerToConcentratedPoolLink{
								{BalancerPoolId: balancerPoolId, ClPoolId: clPool.GetId()},
							},
						},
					)
				}

				// Initialize test incentives on the pool.
				err := clKeeper.SetMultipleIncentiveRecords(s.Ctx, tc.poolIncentiveRecords)
				s.Require().NoError(err)

				// Since the canonical balancer pool automatically claims, ensure that the pool is properly funded if the pool exists.
				if tc.canonicalBalancerPoolAssets != nil {
					for _, incentive := range tc.poolIncentiveRecords {
						s.FundAcc(clPool.GetIncentivesAddress(), sdk.NewCoins(sdk.NewCoin(incentive.IncentiveRecordBody.RemainingCoin.Denom, incentive.IncentiveRecordBody.RemainingCoin.Amount.TruncateInt())))
					}
				}

				// Get initial uptime accum values for comparison
				initUptimeAccumValues, err := clKeeper.GetUptimeAccumulatorValues(s.Ctx, clPool.GetId())
				s.Require().NoError(err)

				// Add qualifying and non-qualifying liquidity to the pool
				qualifyingLiquidity, qualifyingBalancerLiquidity := sdk.ZeroDec(), sdk.ZeroDec()
				if !tc.isInvalidBalancerPool {
					depositedCoins := sdk.NewCoins(sdk.NewCoin(clPool.GetToken0(), testQualifyingDepositsOne), sdk.NewCoin(clPool.GetToken1(), testQualifyingDepositsOne))
					s.FundAcc(testAddressOne, depositedCoins)
					positionData, err := clKeeper.CreatePosition(s.Ctx, clPool.GetId(), testAddressOne, depositedCoins, sdk.ZeroInt(), sdk.ZeroInt(), clPool.GetCurrentTick()-100, clPool.GetCurrentTick()+100)
					s.Require().NoError(err)
					qualifyingLiquidity = positionData.Liquidity

					// If a canonical balancer pool exists, we add its respective shares to the qualifying amount as well.
					clPool, err = clKeeper.GetPoolById(s.Ctx, clPool.GetId())
					s.Require().NoError(err)
					if tc.canonicalBalancerPoolAssets != nil {
						qualifyingBalancerLiquidityPreDiscount := math.GetLiquidityFromAmounts(clPool.GetCurrentSqrtPrice(), types.MinSqrtPrice, types.MaxSqrtPrice, tc.canonicalBalancerPoolAssets[0].Token.Amount, tc.canonicalBalancerPoolAssets[1].Token.Amount)
						qualifyingBalancerLiquidity = (sdk.OneDec().Sub(types.DefaultBalancerSharesDiscount)).Mul(qualifyingBalancerLiquidityPreDiscount)
						qualifyingLiquidity = qualifyingLiquidity.Add(qualifyingBalancerLiquidity)

						actualLiquidityAdded0, actualLiquidityAdded1, err := clPool.CalcActualAmounts(s.Ctx, types.MinInitializedTick, types.MaxTick, qualifyingBalancerLiquidity)
						s.Require().NoError(err)
						s.FundAcc(clPool.GetIncentivesAddress(), sdk.NewCoins(sdk.NewCoin(clPool.GetToken0(), actualLiquidityAdded0.TruncateInt()), sdk.NewCoin(clPool.GetToken1(), actualLiquidityAdded1.TruncateInt())))
					}
				}

				// Let enough time elapse to qualify the position for the first three supported uptimes
				s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(defaultTestUptime))

				// Let `timeElapsed` time pass to test incentive distribution
				s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(tc.timeElapsed))

				// System under test 1
				// Use cache context to avoid persisting updates for the next function
				// that relies on the same test cases and setup.
				cacheCtx, _ := s.Ctx.CacheContext()
				err = clKeeper.UpdatePoolUptimeAccumulatorsToNow(cacheCtx, tc.poolId)

				validateResult(cacheCtx, err, tc, balancerPoolId, clPool.GetId(), initUptimeAccumValues, qualifyingBalancerLiquidity, qualifyingLiquidity)

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

				expectedUptimeDeltas := validateResult(s.Ctx, err, tc, balancerPoolId, clPool.GetId(), initUptimeAccumValues, qualifyingBalancerLiquidity, qualifyingLiquidity)

				if tc.expectedError != nil {
					return
				}

				// Ensure that each uptime accumulater value that was passed in as an argument changes by the correct amount.
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
	defaultInitialLiquidity := sdk.OneDec()
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
				currentTick := pool.GetCurrentTick()

				// Update global uptime accums
				err := addToUptimeAccums(s.Ctx, pool.GetId(), s.App.ConcentratedLiquidityKeeper, tc.globalUptimeGrowth)
				s.Require().NoError(err)

				// Update tick-level uptime trackers
				s.initializeTick(s.Ctx, currentTick, tc.lowerTick, defaultInitialLiquidity, cl.EmptyCoins, wrapUptimeTrackers(tc.lowerTickUptimeGrowthOutside), true)
				s.initializeTick(s.Ctx, currentTick, tc.upperTick, defaultInitialLiquidity, cl.EmptyCoins, wrapUptimeTrackers(tc.upperTickUptimeGrowthOutside), false)
				pool.SetCurrentTick(tc.currentTick)
				err = s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, pool)
				s.Require().NoError(err)

				// system under test
				uptimeGrowthInside, err := s.App.ConcentratedLiquidityKeeper.GetUptimeGrowthInsideRange(s.Ctx, pool.GetId(), tc.lowerTick, tc.upperTick)
				s.Require().NoError(err)

				// check if returned uptime growth inside has correct value
				s.Require().Equal(tc.expectedUptimeGrowthInside, uptimeGrowthInside)

				uptimeGrowthOutside, err := s.App.ConcentratedLiquidityKeeper.GetUptimeGrowthOutsideRange(s.Ctx, pool.GetId(), tc.lowerTick, tc.upperTick)
				s.Require().NoError(err)

				// check if returned uptime growth inside has correct value
				s.Require().Equal(tc.expectedUptimeGrowthOutside, uptimeGrowthOutside)
			})
		}
	})
}

func (s *KeeperTestSuite) TestGetUptimeGrowthErrors() {
	_, err := s.clk.GetUptimeGrowthInsideRange(s.Ctx, defaultPoolId, 0, 0)
	s.Require().Error(err)
	_, err = s.clk.GetUptimeGrowthOutsideRange(s.Ctx, defaultPoolId, 0, 0)
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
		positionLiquidity sdk.Dec

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
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.hundredTokensMultiDenom),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.hundredTokensMultiDenom),
			},
			positionId:               DefaultPositionId,
			currentTickIndex:         0,
			globalUptimeAccumValues:  uptimeHelper.threeHundredTokensMultiDenom,
			expectedInitAccumValue:   uptimeHelper.hundredTokensMultiDenom,
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
		"(lower < curr < upper) non-zero uptime trackers (position already existing)": {
			positionLiquidity: DefaultLiquidityAmt,
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.hundredTokensMultiDenom),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.hundredTokensMultiDenom),
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
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.hundredTokensMultiDenom),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.threeHundredTokensMultiDenom),
			},
			positionId:               DefaultPositionId,
			currentTickIndex:         51,
			globalUptimeAccumValues:  uptimeHelper.fourHundredTokensMultiDenom,
			expectedInitAccumValue:   uptimeHelper.twoHundredTokensMultiDenom,
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
		"(lower < upper < curr) non-zero uptime trackers (position already existing)": {
			positionLiquidity: DefaultLiquidityAmt,
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.hundredTokensMultiDenom),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.threeHundredTokensMultiDenom),
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
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.threeHundredTokensMultiDenom),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.hundredTokensMultiDenom),
			},
			positionId:               DefaultPositionId,
			currentTickIndex:         -51,
			globalUptimeAccumValues:  uptimeHelper.fourHundredTokensMultiDenom,
			expectedInitAccumValue:   uptimeHelper.twoHundredTokensMultiDenom,
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
		"(curr < lower < upper) nonzero uptime trackers (position already existing)": {
			positionLiquidity: DefaultLiquidityAmt,
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.threeHundredTokensMultiDenom),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.hundredTokensMultiDenom),
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
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.varyingTokensMultiDenom),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.hundredTokensMultiDenom),
			},
			positionId:       DefaultPositionId,
			currentTickIndex: 0,

			// We set up the global accum values such that the growth inside is equal to 100 of each denom
			// for each uptime tracker. Let the uptime growth inside (UGI) = 100
			globalUptimeAccumValues: []sdk.DecCoins{
				sdk.NewDecCoins(
					// 100 + 100 + UGI = 300
					sdk.NewDecCoin("bar", sdk.NewInt(300)),
					// 100 + 100 + UGI = 300
					sdk.NewDecCoin("foo", sdk.NewInt(300)),
				),
				sdk.NewDecCoins(
					// 100 + 103 + UGI = 303
					sdk.NewDecCoin("bar", sdk.NewInt(303)),
					// 100 + 101 + UGI = 301
					sdk.NewDecCoin("foo", sdk.NewInt(301)),
				),
				sdk.NewDecCoins(
					// 100 + 106 + UGI = 306
					sdk.NewDecCoin("bar", sdk.NewInt(306)),
					// 100 + 102 + UGI = 302
					sdk.NewDecCoin("foo", sdk.NewInt(302)),
				),
				sdk.NewDecCoins(
					// 100 + 109 + UGI = 309
					sdk.NewDecCoin("bar", sdk.NewInt(309)),
					// 100 + 103 + UGI = 303
					sdk.NewDecCoin("foo", sdk.NewInt(303)),
				),
				sdk.NewDecCoins(
					// 100 + 112 + UGI = 312
					sdk.NewDecCoin("bar", sdk.NewInt(312)),
					// 100 + 104 + UGI = 304
					sdk.NewDecCoin("foo", sdk.NewInt(304)),
				),
				sdk.NewDecCoins(
					// 100 + 115 + UGI = 315
					sdk.NewDecCoin("bar", sdk.NewInt(315)),
					// 100 + 105 + UGI = 305
					sdk.NewDecCoin("foo", sdk.NewInt(305)),
				),
			},
			// Equal to 100 of foo and bar in each uptime tracker (UGI)
			expectedInitAccumValue:   uptimeHelper.hundredTokensMultiDenom,
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
		"error: negative liquidity for first position": {
			positionLiquidity: DefaultLiquidityAmt.Neg(),
			lowerTick: tick{
				tickIndex:      -50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.hundredTokensMultiDenom),
			},
			upperTick: tick{
				tickIndex:      50,
				uptimeTrackers: wrapUptimeTrackers(uptimeHelper.hundredTokensMultiDenom),
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

				// Set blocktime to fixed UTC value for consistency
				s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

				clPool := s.PrepareConcentratedPool()

				// Initialize lower, upper, and current ticks
				s.initializeTick(s.Ctx, test.currentTickIndex, test.lowerTick.tickIndex, sdk.ZeroDec(), cl.EmptyCoins, test.lowerTick.uptimeTrackers, true)
				s.initializeTick(s.Ctx, test.currentTickIndex, test.upperTick.tickIndex, sdk.ZeroDec(), cl.EmptyCoins, test.upperTick.uptimeTrackers, false)
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

					s.initializeTick(s.Ctx, test.currentTickIndex, test.newLowerTick.tickIndex, sdk.ZeroDec(), cl.EmptyCoins, test.newLowerTick.uptimeTrackers, true)
					s.initializeTick(s.Ctx, test.currentTickIndex, test.newUpperTick.tickIndex, sdk.ZeroDec(), cl.EmptyCoins, test.newUpperTick.uptimeTrackers, false)
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
						s.Require().Equal(sdk.NewDec(2).Mul(test.positionLiquidity), positionRecord.NumShares)
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
		liquidity   sdk.Dec
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
		existingAccumLiquidity   []sdk.Dec
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
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)),
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
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)),
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
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)),
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
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)),
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
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)),
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
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)),
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
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)),
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
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)),
		},
		"other liquidity on uptime accums: (lower < curr < upper) uptime growth both inside and outside range, 1D time in position": {
			currentTick: 1,
			existingAccumLiquidity: []sdk.Dec{
				sdk.NewDec(99900123432),
				sdk.NewDec(18942),
				sdk.NewDec(0),
				sdk.NewDec(9981),
				sdk.NewDec(1),
				sdk.NewDec(778212931834),
			},
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default0To2PosParam,
			numPositions:             1,
			timeInPosition:           oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)),
		},
		"multiple positions in same range: (lower < curr < upper) uptime growth both inside and outside range, 1D time in position": {
			currentTick: 1,
			existingAccumLiquidity: []sdk.Dec{
				sdk.NewDec(99900123432),
				sdk.NewDec(18942),
				sdk.NewDec(0),
				sdk.NewDec(9981),
				sdk.NewDec(1),
				sdk.NewDec(778212931834),
			},
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams:           default0To2PosParam,
			numPositions:             3,
			timeInPosition:           oneDay,
			// Since we introduced positionIDs, despite these position having the same range and pool, only
			// the position ID being claimed will be considered for the claim.
			expectedIncentivesClaimed:   expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
			expectedForfeitedIncentives: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier)),
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
					s.initializeTick(s.Ctx, tc.currentTick, tc.positionParams.lowerTick, tc.positionParams.liquidity, cl.EmptyCoins, wrapUptimeTrackers(uptimeHelper.emptyExpectedAccumValues), true)
					s.initializeTick(s.Ctx, tc.currentTick, tc.positionParams.upperTick, tc.positionParams.liquidity, cl.EmptyCoins, wrapUptimeTrackers(uptimeHelper.emptyExpectedAccumValues), false)

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
						s.addUptimeGrowthInsideRange(s.Ctx, validPoolId, ownerWithValidPosition, tc.currentTick, tc.positionParams.lowerTick, tc.positionParams.upperTick, tc.addedUptimeGrowthInside)
					}

					// Add to uptime growth outside range
					if tc.addedUptimeGrowthOutside != nil {
						s.addUptimeGrowthOutsideRange(s.Ctx, validPoolId, ownerWithValidPosition, tc.currentTick, tc.positionParams.lowerTick, tc.positionParams.upperTick, tc.addedUptimeGrowthOutside)
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
				actualIncentivesClaimed, actualIncetivesForfeited, err := clKeeper.CollectIncentives(s.Ctx, ownerWithValidPosition, DefaultPositionId)

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
				s.Require().Equal(tc.expectedIncentivesClaimed.Add(tc.expectedForfeitedIncentives...).String(), (incentivesBalanceBeforeCollect.Sub(incentivesBalanceAfterCollect)).String())
				s.Require().Equal(tc.expectedIncentivesClaimed.String(), (ownerBalancerAfterCollect.Sub(ownerBalancerBeforeCollect)).String())
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
		sender                   sdk.AccAddress
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
					sdk.NewInt(8),
				),
			),
			recordToSet: withAmount(incentiveRecordOne, sdk.NewDec(8)),
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
					sdk.ZeroInt(),
				),
			),
			recordToSet: withAmount(incentiveRecordOne, sdk.ZeroDec()),

			expectedError: types.InvalidIncentiveCoinError{PoolId: 1, IncentiveCoin: sdk.NewCoin(incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, sdk.ZeroInt())},
		},
		"invalid incentive coin (negative)": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					sdk.ZeroInt(),
				),
			),
			recordToSet:              withAmount(incentiveRecordOne, sdk.ZeroDec()),
			useNegativeIncentiveCoin: true,

			expectedError: types.InvalidIncentiveCoinError{PoolId: 1, IncentiveCoin: sdk.Coin{Denom: incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom, Amount: sdk.NewInt(-1)}},
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
			recordToSet: withEmissionRate(incentiveRecordOne, sdk.ZeroDec()),

			expectedError: types.NonPositiveEmissionRateError{PoolId: 1, EmissionRate: sdk.ZeroDec()},
		},
		"negative emission rate": {
			poolId: defaultPoolId,
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Denom,
					incentiveRecordOne.IncentiveRecordBody.RemainingCoin.Amount.Ceil().RoundInt(),
				),
			),
			recordToSet: withEmissionRate(incentiveRecordOne, sdk.NewDec(-1)),

			expectedError: types.NonPositiveEmissionRateError{PoolId: 1, EmissionRate: sdk.NewDec(-1)},
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
				s.FundAcc(tc.sender, tc.senderBalance)

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
					incentiveCoin = sdk.Coin{Denom: tc.recordToSet.IncentiveRecordBody.RemainingCoin.Denom, Amount: sdk.NewInt(-1)}
				}

				// Set the next incentive record id.
				originalNextIncentiveRecordId := tc.recordToSet.IncentiveId
				clKeeper.SetNextIncentiveRecordId(s.Ctx, originalNextIncentiveRecordId)

				// system under test
				incentiveRecord, err := clKeeper.CreateIncentive(s.Ctx, tc.poolId, tc.sender, incentiveCoin, tc.recordToSet.IncentiveRecordBody.EmissionRate, tc.recordToSet.IncentiveRecordBody.StartTime, tc.recordToSet.MinUptime)

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
		incentiveCoin = sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000000000))
		emissionRate  = sdk.NewDecWithPrec(1, 2)
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
				err := poolSpreadRewardsAccumulator.NewPosition(positionKey, sdk.OneDec(), nil)
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

// Note that the non-forfeit cases are thoroughly tested in `TestCollectIncentives`
func (s *KeeperTestSuite) TestQueryAndClaimAllIncentives() {
	uptimeHelper := getExpectedUptimes()
	defaultSender := s.TestAccs[0]
	tests := map[string]struct {
		numShares         sdk.Dec
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
			numShares:        sdk.OneDec(),
		},
		"claim and forfeit rewards (2 shares)": {
			positionIdCreate:  DefaultPositionId,
			positionIdClaim:   DefaultPositionId,
			defaultJoinTime:   true,
			growthInside:      uptimeHelper.hundredTokensMultiDenom,
			growthOutside:     uptimeHelper.twoHundredTokensMultiDenom,
			forfeitIncentives: true,
			numShares:         sdk.NewDec(2),
		},
		"claim and forfeit rewards when no rewards have accrued": {
			positionIdCreate:  DefaultPositionId,
			positionIdClaim:   DefaultPositionId,
			defaultJoinTime:   true,
			forfeitIncentives: true,
			numShares:         sdk.OneDec(),
		},
		"claim and forfeit rewards with varying amounts and different denoms": {
			positionIdCreate:  DefaultPositionId,
			positionIdClaim:   DefaultPositionId,
			defaultJoinTime:   true,
			growthInside:      uptimeHelper.varyingTokensMultiDenom,
			growthOutside:     uptimeHelper.varyingTokensSingleDenom,
			forfeitIncentives: true,
			numShares:         sdk.OneDec(),
		},

		// error catching

		"error: non existent position": {
			positionIdCreate: DefaultPositionId,
			positionIdClaim:  DefaultPositionId + 1, // non existent position
			defaultJoinTime:  true,
			growthInside:     uptimeHelper.hundredTokensMultiDenom,
			growthOutside:    uptimeHelper.twoHundredTokensMultiDenom,
			numShares:        sdk.OneDec(),

			expectedError: types.PositionIdNotFoundError{PositionId: DefaultPositionId + 1},
		},

		"error: negative duration": {
			positionIdCreate: DefaultPositionId,
			positionIdClaim:  DefaultPositionId,
			defaultJoinTime:  false,
			growthInside:     uptimeHelper.hundredTokensMultiDenom,
			growthOutside:    uptimeHelper.twoHundredTokensMultiDenom,
			numShares:        sdk.OneDec(),

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
				err := s.clk.InitOrUpdatePosition(s.Ctx, validPoolId, defaultSender, DefaultLowerTick, DefaultUpperTick, tc.numShares, joinTime, tc.positionIdCreate)
				s.Require().NoError(err)

				clPool.SetCurrentTick(DefaultCurrTick)
				if tc.growthOutside != nil {
					s.addUptimeGrowthOutsideRange(s.Ctx, validPoolId, defaultSender, DefaultCurrTick, DefaultLowerTick, DefaultUpperTick, tc.growthOutside)
				}

				if tc.growthInside != nil {
					s.addUptimeGrowthInsideRange(s.Ctx, validPoolId, defaultSender, DefaultCurrTick, DefaultLowerTick, DefaultUpperTick, tc.growthInside)
				}

				err = s.clk.SetPool(s.Ctx, clPool)
				s.Require().NoError(err)

				preCommunityPoolBalance := bankKeeper.GetAllBalances(s.Ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName))

				// Store initial pool and sender balances for comparison later
				initSenderBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, defaultSender)
				initPoolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, clPool.GetAddress())

				largestSupportedUptime := s.clk.GetLargestSupportedUptimeDuration(s.Ctx)
				if !tc.forfeitIncentives {
					// Let enough time elapse for the position to accrue rewards for all supported uptimes.
					s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(largestSupportedUptime))
				}

				// --- System under test ---
				amountClaimedQuery, amountForfeitedQuery, err := s.clk.GetClaimableIncentives(s.Ctx, tc.positionIdClaim)

				// Pull new balances for comparison
				newSenderBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, defaultSender)
				newPoolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, clPool.GetAddress())

				if tc.expectedError != nil {
					s.Require().ErrorIs(err, tc.expectedError)
				}

				// Ensure balances have not been mutated (since this is a query)
				s.Require().Equal(initSenderBalances, newSenderBalances)
				s.Require().Equal(initPoolBalances, newPoolBalances)

				amountClaimed, amountForfeited, err := s.clk.PrepareClaimAllIncentivesForPosition(s.Ctx, tc.positionIdClaim)

				// --- Assertions ---

				// Pull new balances for comparison
				newSenderBalances = s.App.BankKeeper.GetAllBalances(s.Ctx, defaultSender)
				newPoolBalances = s.App.BankKeeper.GetAllBalances(s.Ctx, clPool.GetAddress())

				s.Require().Equal(amountClaimedQuery, amountClaimed)
				s.Require().Equal(amountForfeitedQuery, amountForfeited)

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
					communityPoolBalanceDelta := postCommunityPoolBalance.Sub(preCommunityPoolBalance)
					s.Require().Equal(sdk.Coins{}, amountClaimed)
					s.Require().Equal("", communityPoolBalanceDelta.String())
				} else {
					// We expect claimed rewards to be equal to growth inside
					expectedCoins := sdk.Coins(nil)
					for _, growthInside := range tc.growthInside {
						expectedCoins = expectedCoins.Add(sdk.NormalizeCoins(growthInside)...)
					}
					s.Require().Equal(expectedCoins, amountClaimed)
					s.Require().Equal(sdk.Coins{}, amountForfeited)
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
			expectedCoins:            sdk.NewCoins(sdk.NewCoin(USDC, sdk.NewInt(59))), //  after 1min = 59.999999999901820104usdc ~ 59usdc becasue 1usdc emitted every second
			minUptimeIncentiveRecord: time.Nanosecond,
		},
		{
			name:                     "Claim after 1 hr, 1ns uptime",
			blockTimeElapsed:         time.Hour,
			expectedCoins:            sdk.NewCoins(sdk.NewCoin(USDC, sdk.NewInt(3599))), //  after 1min = 59.999999999901820104usdc ~ 59usdc becasue 1usdc emitted every second
			minUptimeIncentiveRecord: time.Nanosecond,
		},
		{
			name:                     "Claim after 24 hours, 1ns uptime",
			blockTimeElapsed:         time.Hour * 24,
			expectedCoins:            sdk.NewCoins(sdk.NewCoin(USDC, sdk.NewInt(9999))), //  after 24hr > 2hr46min = 9999usdc.999999999901820104 ~ 9999usdc
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
			expectedCoins:            sdk.NewCoins(sdk.NewCoin(USDC, sdk.NewInt(9999))), //  after 24hr > 2hr46min = 9999usdc.999999999901820104 ~ 9999usdc
			minUptimeIncentiveRecord: time.Hour * 24,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Init suite for the test.
			s.SetupTest()

			requiredBalances := sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(1_000_000)), sdk.NewCoin(USDC, sdk.NewInt(5_000_000_000)))
			s.FundAcc(s.TestAccs[0], requiredBalances)
			s.FundAcc(s.TestAccs[1], requiredBalances)

			// Create CL pool
			pool := s.PrepareConcentratedPool()

			// Set up position
			positionOneData, err := s.clk.CreatePosition(s.Ctx, pool.GetId(), s.TestAccs[0], requiredBalances, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
			s.Require().NoError(err)

			// Set incentives for pool to ensure accumulators work correctly
			testIncentiveRecord := types.IncentiveRecord{
				PoolId: pool.GetId(),
				IncentiveRecordBody: types.IncentiveRecordBody{
					RemainingCoin: sdk.NewDecCoinFromDec(USDC, sdk.NewDec(10_000)), // 2hr 46m to emit all incentives
					EmissionRate:  sdk.NewDec(1),                                   // 1 per second
					StartTime:     s.Ctx.BlockTime(),
				},
				MinUptime: tc.minUptimeIncentiveRecord,
			}
			err = s.clk.SetMultipleIncentiveRecords(s.Ctx, []types.IncentiveRecord{testIncentiveRecord})
			s.Require().NoError(err)

			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(tc.blockTimeElapsed))

			// Update the uptime accumulators to the current block time.
			// This is done to determine the exact amount of incentives we expect to be forfeited, if any.
			err = s.clk.UpdatePoolUptimeAccumulatorsToNow(s.Ctx, pool.GetId())
			s.Require().NoError(err)

			// Retrieve the uptime accumulators for the pool.
			uptimeAccumulatorsPreClaim, err := s.clk.GetUptimeAccumulators(s.Ctx, pool.GetId())
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
						collectedIncentivesForUptime, _ := outstandingRewards.TruncateDecimal()

						for _, coin := range collectedIncentivesForUptime {
							expectedForfeitedIncentives = expectedForfeitedIncentives.Add(coin)
						}
					}
				}
			}

			// System under test
			collectedInc, forfeitedIncentives, err := s.clk.PrepareClaimAllIncentivesForPosition(s.Ctx, positionOneData.ID)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedCoins.String(), collectedInc.String())
			s.Require().Equal(expectedForfeitedIncentives.String(), forfeitedIncentives.String())

			// The difference accumulator value should have increased if we forfeited incentives by claiming.
			uptimeAccumsDiffPostClaim := sdk.NewDecCoins()
			if tc.blockTimeElapsed < tc.minUptimeIncentiveRecord {
				uptimeAccumulatorsPostClaim, err := s.clk.GetUptimeAccumulators(s.Ctx, pool.GetId())
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
		clParams := s.clk.GetParams(s.Ctx)
		clParams.AuthorizedUptimes = []time.Duration{time.Nanosecond, testFullChargeDuration}
		s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, clParams)

		// Fund accounts twice because two positions are created.
		s.FundAcc(defaultAddress, requiredBalances.Add(requiredBalances...))

		// Create CL pool
		pool := s.PrepareConcentratedPool()

		expectedAmount := sdk.NewInt(60 * 60 * 24) // 1 day in seconds * 1 per second

		oneUUSDCCoin := sdk.NewCoin(USDC, sdk.OneInt())
		// -1 for acceptable rounding error
		expectedCoinsPerFullCharge := sdk.NewCoins(sdk.NewCoin(USDC, expectedAmount.Sub(sdk.OneInt())))
		expectedHalfOfExpectedCoinsPerFullCharge := sdk.NewCoins(sdk.NewCoin(USDC, expectedAmount.QuoRaw(2).Sub(sdk.OneInt())))

		// Multiplied by 3 because we change the block time 3 times and claim
		// 1. by directly calling CollectIncentives
		// 2. by calling WithdrawPosition
		// 3. by calling CollectIncentives
		s.FundAcc(pool.GetIncentivesAddress(), sdk.NewCoins(sdk.NewCoin(USDC, expectedAmount.Mul(sdk.NewInt(3)))))
		// Set incentives for pool to ensure accumulators work correctly
		testIncentiveRecord := types.IncentiveRecord{
			PoolId: 1,
			IncentiveRecordBody: types.IncentiveRecordBody{
				RemainingCoin: sdk.NewDecCoinFromDec(USDC, sdk.NewDec(1000000000000000000)),
				EmissionRate:  sdk.NewDec(1), // 1 per second
				StartTime:     defaultBlockTime,
			},
			MinUptime: time.Nanosecond,
		}
		err := s.App.ConcentratedLiquidityKeeper.SetMultipleIncentiveRecords(s.Ctx, []types.IncentiveRecord{testIncentiveRecord})
		s.Require().NoError(err)

		// Set up position
		positionOneData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, defaultPoolId, defaultAddress, DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
		s.Require().NoError(err)

		// Increase block time by the fully charged duration (first time)
		s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(testFullChargeDuration))

		// Claim incentives.
		collected, _, err := s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, defaultAddress, positionOneData.ID)
		s.Require().NoError(err)
		s.Require().Equal(expectedCoinsPerFullCharge.String(), collected.String())

		// Increase block time by the fully charged duration (second time)
		s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(testFullChargeDuration))

		// Create another position
		positionTwoData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, defaultPoolId, defaultAddress, DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
		s.Require().NoError(err)

		// Increase block time by the fully charged duration (third time)
		s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(testFullChargeDuration))

		// Claim for second position. Must only claim half of the original expected amount since now there are 2 positions.
		collected, _, err = s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, defaultAddress, positionTwoData.ID)
		s.Require().NoError(err)
		s.Require().Equal(expectedHalfOfExpectedCoinsPerFullCharge.String(), collected.String())

		// Claim for first position and observe that claims full expected charge for the period between 1st claim and 2nd position creation
		// and half of the full charge amount since the 2nd position was created.
		collected, _, err = s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, defaultAddress, positionOneData.ID)
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

// TestPrepareBalancerPoolAsFullRange tests the preparation of a balancer pool as a full range position on its linked concentrated pool.
// Note that regardless of which uptimes are authorized, we always add records on all uptime accumulators so this test simply runs checks
// on pool-wide supported uptimes.
func (s *KeeperTestSuite) TestPrepareBalancerPoolAsFullRange() {
	defaultBalancerAssets := []balancer.PoolAsset{
		{Weight: sdk.NewInt(1), Token: sdk.NewCoin("foo", sdk.NewInt(1000000000))},
		{Weight: sdk.NewInt(1), Token: sdk.NewCoin("bar", sdk.NewInt(1000000000))},
	}
	defaultConcentratedAssets := sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(100)), sdk.NewCoin("bar", sdk.NewInt(100)))

	type testcase struct {
		// defaults to defaultConcentratedAssets
		existingConcentratedLiquidity sdk.Coins
		// defaults to defaultBalancerAssets
		balancerPoolAssets []balancer.PoolAsset
		// defaults sdk.OneDec()
		portionOfSharesBonded        sdk.Dec
		balancerSharesRewardDiscount sdk.Dec

		noCanonicalBalancerPool bool
		flipAsset0And1          bool
		expectedError           error
	}
	initTestCase := func(tc testcase) testcase {
		if tc.existingConcentratedLiquidity.Empty() {
			tc.existingConcentratedLiquidity = defaultConcentratedAssets
		}
		if len(tc.balancerPoolAssets) == 0 {
			tc.balancerPoolAssets = defaultBalancerAssets
		}
		if (tc.portionOfSharesBonded == sdk.Dec{}) {
			tc.portionOfSharesBonded = sdk.NewDec(1)
		}
		if (tc.balancerSharesRewardDiscount == sdk.Dec{}) {
			tc.balancerSharesRewardDiscount = sdk.ZeroDec()
		}
		return tc
	}

	tests := map[string]testcase{
		"happy path: balancer and CL pool at same spot price": {
			// 100 existing shares and 100 shares added from balancer
			// no other test will show defaults.
			existingConcentratedLiquidity: defaultConcentratedAssets,
			balancerPoolAssets:            defaultBalancerAssets,
			portionOfSharesBonded:         sdk.NewDec(1),
		},
		"happy path: balancer and CL pool at same spot price, flip assets to test secondary logic branch": {
			flipAsset0And1:        true,
			balancerPoolAssets:    defaultBalancerAssets,
			portionOfSharesBonded: sdk.NewDec(1),
		},
		"same spot price, different total share amount": {
			// 100 existing shares and 200 shares added from balancer
			balancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("foo", sdk.NewInt(200))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("bar", sdk.NewInt(200))},
			},
		},
		"different spot price between balancer and CL pools (excess asset0)": {
			// 100 existing shares and 100 shares added from balancer. We expect only the even portion of
			// the Balancer pool to be joined, with the remaining 50foo not qualifying.
			balancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("foo", sdk.NewInt(150))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("bar", sdk.NewInt(100))},
			},
		},
		"different spot price between balancer and CL pools (excess asset1)": {
			// 100 existing shares and 100 shares added from balancer. We expect only the even portion of
			// the Balancer pool to be joined, with the remaining 50bar not qualifying.
			balancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("foo", sdk.NewInt(100))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("bar", sdk.NewInt(150))},
			},
		},
		"same spot price, different total share amount, only half bonded": {
			// 100 existing shares and 200 shares added from balancer
			balancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("foo", sdk.NewInt(200))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("bar", sdk.NewInt(200))},
			},
			portionOfSharesBonded: sdk.MustNewDecFromStr("0.5"),
		},
		"different spot price between balancer and CL pools (excess asset1), only partially bonded": {
			// 100 existing shares and 100 shares added from balancer. We expect only the even portion of
			// the Balancer pool to be joined, with the remaining 50bar not qualifying.
			balancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("foo", sdk.NewInt(100))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("bar", sdk.NewInt(150))},
			},
			portionOfSharesBonded: sdk.MustNewDecFromStr("0.1"),
		},
		"no canonical balancer pool": {
			// 100 existing shares and 100 shares added from balancer
			// Note that we expect this to fail quietly, as most CL pools will not have linked Balancer pools
			noCanonicalBalancerPool: true,
		},
	}

	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			s.Run(name, func() {
				// --- Setup test env ---
				s.SetupTest()
				tc = initTestCase(tc)

				balancerPoolId := s.setupBalancerPoolWithFractionLocked(tc.balancerPoolAssets, tc.portionOfSharesBonded)

				var clPool types.ConcentratedPoolExtension
				if tc.flipAsset0And1 {
					clPool = s.PrepareCustomConcentratedPool(s.TestAccs[0], tc.existingConcentratedLiquidity[1].Denom, tc.existingConcentratedLiquidity[0].Denom, DefaultTickSpacing, sdk.ZeroDec())
				} else {
					clPool = s.PrepareCustomConcentratedPool(s.TestAccs[0], tc.existingConcentratedLiquidity[0].Denom, tc.existingConcentratedLiquidity[1].Denom, DefaultTickSpacing, sdk.ZeroDec())
				}

				// Set up an existing full range position. Note that the second return value is the position ID, not an error.
				initialLiquidity, _ := s.SetupPosition(clPool.GetId(), s.TestAccs[0], tc.existingConcentratedLiquidity, DefaultMinTick, DefaultMaxTick, false)

				if tc.noCanonicalBalancerPool {
					balancerPoolId = 0
				} else {
					s.App.GAMMKeeper.OverwriteMigrationRecordsAndRedirectDistrRecords(s.Ctx,
						gammmigration.MigrationRecords{
							BalancerToConcentratedPoolLinks: []gammmigration.BalancerToConcentratedPoolLink{
								{BalancerPoolId: balancerPoolId, ClPoolId: clPool.GetId()},
							},
						},
					)
				}

				// Calculate balancer share amount for full range
				updatedClPool, err := s.clk.GetPoolById(s.Ctx, clPool.GetId())
				s.Require().NoError(err)
				asset0BalancerAmount := tc.balancerPoolAssets[0].Token.Amount.ToDec().Mul(tc.portionOfSharesBonded).TruncateInt()
				asset1BalancerAmount := tc.balancerPoolAssets[1].Token.Amount.ToDec().Mul(tc.portionOfSharesBonded).TruncateInt()
				qualifyingSharesPreDiscount := math.GetLiquidityFromAmounts(updatedClPool.GetCurrentSqrtPrice(), types.MinSqrtPrice, types.MaxSqrtPrice, asset1BalancerAmount, asset0BalancerAmount)
				qualifyingShares := (sdk.OneDec().Sub(types.DefaultBalancerSharesDiscount)).Mul(qualifyingSharesPreDiscount)

				// TODO: clean this check up (will likely require refactoring the whole test)
				clearOutQualifyingShares := tc.noCanonicalBalancerPool
				if clearOutQualifyingShares {
					qualifyingShares = sdk.NewDec(0)
				}

				// --- System under test ---

				// Get uptime accums for the cl pool.
				uptimeAccums, err := s.clk.GetUptimeAccumulators(s.Ctx, clPool.GetId())
				s.Require().NoError(err)

				retrievedBalancerPoolId, addedLiquidity, err := s.clk.PrepareBalancerPoolAsFullRange(s.Ctx, clPool.GetId(), uptimeAccums)

				// --- Assertions ---

				s.Require().NoError(err)
				// Ensure that returned balancer pool ID is correct
				s.Require().Equal(balancerPoolId, retrievedBalancerPoolId)

				// General assertions regardless of error
				updatedClPool, err = s.clk.GetPoolById(s.Ctx, clPool.GetId())
				s.Require().NoError(err)

				clPoolUptimeAccumulatorsFromState, err := s.clk.GetUptimeAccumulators(s.Ctx, clPool.GetId())
				s.Require().NoError(err)

				s.Require().True(len(clPoolUptimeAccumulatorsFromState) > 0)
				expectedShares := qualifyingShares.Add(initialLiquidity)
				for uptimeIdx, uptimeAccum := range clPoolUptimeAccumulatorsFromState {
					currAccumShares := uptimeAccum.GetTotalShares()

					// Ensure each accum has the correct number of final shares
					s.Require().Equal(expectedShares, currAccumShares)

					// Also validate uptime accumulators passed in as parameter.
					currAccumShares = uptimeAccums[uptimeIdx].GetTotalShares()
					s.Require().Equal(expectedShares, currAccumShares)
				}

				// Ensure added liquidity is equal to the amount accum shares changed by
				s.Require().Equal(qualifyingShares, addedLiquidity)

				// Pool liquidity should remain unchanged
				s.Require().Equal(initialLiquidity, updatedClPool.GetLiquidity())
			})
		}
	})
}

func (s *KeeperTestSuite) TestPrepareBalancerPoolAsFullRangeWithNonExistentPools() {
	existingConcentratedAssets := sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(100)), sdk.NewCoin("bar", sdk.NewInt(100)))

	type testcase struct {
		// defaults to defaultBalancerAssets
		balancerPoolAssets []balancer.PoolAsset

		noBalancerPoolWithID      bool
		invalidConcentratedPoolID bool

		expectedError error
	}
	tests := map[string]testcase{
		// Error catching
		"canonical balancer pool ID exists but pool itself is not found": {
			// 100 existing shares and 100 shares added from balancer
			noBalancerPoolWithID: true,
			expectedError:        gammtypes.PoolDoesNotExistError{PoolId: invalidPoolId},
		},
		"canonical balancer pool has invalid number of assets": {
			// 100 existing shares and 100 shares added from balancer
			balancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("foo", sdk.NewInt(100))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("bar", sdk.NewInt(100))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("baz", sdk.NewInt(100))},
			},
			expectedError: types.ErrInvalidBalancerPoolLiquidityError{ClPoolId: 1, BalancerPoolId: 2, BalancerPoolLiquidity: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(100)), sdk.NewCoin("bar", sdk.NewInt(100)), sdk.NewCoin("baz", sdk.NewInt(100)))},
		},
		"invalid concentrated pool ID": {
			// 100 existing shares and 100 shares added from balancer
			invalidConcentratedPoolID: true,
			expectedError:             types.PoolNotFoundError{PoolId: invalidPoolId + 1},
		},
	}
	// create invalid denom test cases. Either denom1, denom2 or both are invalid
	denomSelector := [][]string{{"foo", "invalid1"}, {"bar", "invalid2"}}
	for i := 0; i < 2; i++ {
		pa1 := balancer.PoolAsset{Weight: sdk.NewInt(1), Token: sdk.NewCoin(denomSelector[0][i], sdk.NewInt(100))}
		for j := 1 - i; j < 2; j++ {
			pa2 := balancer.PoolAsset{Weight: sdk.NewInt(1), Token: sdk.NewCoin(denomSelector[1][j], sdk.NewInt(100))}
			testname := fmt.Sprintf("canonical balancer pool; denom1_invalid=%v; denom2_invalid=%v", i == 1, j == 1)
			tests[testname] = testcase{
				balancerPoolAssets: []balancer.PoolAsset{pa1, pa2},
				expectedError:      types.ErrInvalidBalancerPoolLiquidityError{ClPoolId: 1, BalancerPoolId: 2, BalancerPoolLiquidity: sdk.NewCoins(pa1.Token, pa2.Token)},
			}
		}
	}
	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			s.Run(name, func() {
				s.SetupTest()
				if len(tc.balancerPoolAssets) == 0 {
					tc.balancerPoolAssets = defaultBalancerAssets
				}
				clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], existingConcentratedAssets[0].Denom, existingConcentratedAssets[1].Denom, DefaultTickSpacing, sdk.ZeroDec())

				// Set up an existing full range position. Note that the second return value is the position ID, not an error.
				s.SetupPosition(clPool.GetId(), s.TestAccs[0], existingConcentratedAssets, DefaultMinTick, DefaultMaxTick, false)

				// If a canonical balancer pool exists, we create it and link it with the CL pool
				balancerPoolId := s.setupBalancerPoolWithFractionLocked(tc.balancerPoolAssets, sdk.OneDec())

				if tc.noBalancerPoolWithID {
					balancerPoolId = invalidPoolId
				}

				s.App.GAMMKeeper.OverwriteMigrationRecordsAndRedirectDistrRecords(s.Ctx,
					gammmigration.MigrationRecords{
						BalancerToConcentratedPoolLinks: []gammmigration.BalancerToConcentratedPoolLink{
							{BalancerPoolId: balancerPoolId, ClPoolId: clPool.GetId()},
						},
					},
				)

				concentratedPoolId := clPool.GetId()
				if tc.invalidConcentratedPoolID {
					concentratedPoolId = invalidPoolId + 1
				}

				// Get uptime accums for the cl pool.
				uptimeAccums, err := s.clk.GetUptimeAccumulators(s.Ctx, clPool.GetId())
				s.Require().NoError(err)

				retrievedBalancerPoolId, _, err := s.clk.PrepareBalancerPoolAsFullRange(s.Ctx, concentratedPoolId, uptimeAccums)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				s.Require().Equal(uint64(0), retrievedBalancerPoolId)
			})
		}
	})
}

func (s *KeeperTestSuite) TestClaimAndResetFullRangeBalancerPool() {
	longestLockableDuration, err := s.App.PoolIncentivesKeeper.GetLongestLockableDuration(s.Ctx)
	s.Require().NoError(err)
	uptimeHelper := getExpectedUptimes()

	tests := map[string]struct {
		existingConcentratedLiquidity sdk.Coins
		balancerPoolAssets            []balancer.PoolAsset
		uptimeGrowth                  []sdk.DecCoins

		concentratedPoolDoesNotExist bool
		balancerPoolDoesNotExist     bool
		balSharesNotAddedToAccums    bool
		insufficientPoolBalance      bool

		expectedError error
	}{
		"happy path: valid CL and bal pool IDs": {
			// 100 existing shares and 100 shares added from balancer
			existingConcentratedLiquidity: defaultConcentratedAssets,
			balancerPoolAssets:            defaultBalancerAssets,
			uptimeGrowth:                  uptimeHelper.hundredTokensMultiDenom,
		},
		"valid pool IDs with no uptime growth": {
			// 100 existing shares and 100 shares added from balancer
			existingConcentratedLiquidity: defaultConcentratedAssets,
			balancerPoolAssets:            defaultBalancerAssets,
			uptimeGrowth:                  uptimeHelper.emptyExpectedAccumValues,
		},
		"valid pool IDs with uneven uptime growth": {
			// 100 existing shares and 100 shares added from balancer
			existingConcentratedLiquidity: defaultConcentratedAssets,
			balancerPoolAssets:            defaultBalancerAssets,
			uptimeGrowth:                  uptimeHelper.varyingTokensMultiDenom,
		},
		"different liquidity amounts between balancer and CL pools": {
			// 100 existing shares and 200 shares added from balancer
			existingConcentratedLiquidity: defaultConcentratedAssets,
			balancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("foo", sdk.NewInt(200))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("bar", sdk.NewInt(200))},
			},
			uptimeGrowth: uptimeHelper.emptyExpectedAccumValues,
		},
		"balancer spot price different than CL spot price (foo higher)": {
			// 100 existing shares and 200 shares added from balancer
			// Note that only 200foo/200bar qualify, and the remaining 50bar is not counted
			existingConcentratedLiquidity: defaultConcentratedAssets,
			balancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("foo", sdk.NewInt(200))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("bar", sdk.NewInt(250))},
			},
			uptimeGrowth: uptimeHelper.emptyExpectedAccumValues,
		},
		"balancer spot price different than CL spot price (bar higher)": {
			// 100 existing shares and 200 shares added from balancer
			// Note that only 200foo/200bar qualify, and the remaining 50foo is not counted
			existingConcentratedLiquidity: defaultConcentratedAssets,
			balancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("foo", sdk.NewInt(250))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("bar", sdk.NewInt(200))},
			},
			uptimeGrowth: uptimeHelper.emptyExpectedAccumValues,
		},
		"rounding check: large and imbalanced CL amounts": {
			existingConcentratedLiquidity: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(2<<60)), sdk.NewCoin("bar", sdk.NewInt(2<<61))),
			balancerPoolAssets:            defaultBalancerAssets,
			uptimeGrowth:                  uptimeHelper.hundredTokensMultiDenom,
		},
		"rounding check: large and imbalanced balancer amounts": {
			existingConcentratedLiquidity: defaultConcentratedAssets,
			balancerPoolAssets: []balancer.PoolAsset{
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("foo", sdk.NewInt(2<<61))},
				{Weight: sdk.NewInt(1), Token: sdk.NewCoin("bar", sdk.NewInt(2<<60))},
			},
			uptimeGrowth: uptimeHelper.hundredTokensMultiDenom,
		},

		// Error catching

		"CL pool does not exist": {
			// 100 existing shares and 100 shares added from balancer
			existingConcentratedLiquidity: defaultConcentratedAssets,
			balancerPoolAssets:            defaultBalancerAssets,
			uptimeGrowth:                  uptimeHelper.hundredTokensMultiDenom,

			concentratedPoolDoesNotExist: true,
			expectedError:                types.PoolNotFoundError{PoolId: invalidPoolId + 1},
		},
		"Balancer pool does not exist": {
			// 100 existing shares and 100 shares added from balancer
			existingConcentratedLiquidity: defaultConcentratedAssets,
			balancerPoolAssets:            defaultBalancerAssets,
			uptimeGrowth:                  uptimeHelper.hundredTokensMultiDenom,

			balancerPoolDoesNotExist: true,
			expectedError:            poolincentivestypes.NoGaugeAssociatedWithPoolError{PoolId: invalidPoolId, Duration: longestLockableDuration},
		},
		"Balancer shares not yet added to CL pool accums": {
			// 100 existing shares and 100 shares added from balancer
			existingConcentratedLiquidity: defaultConcentratedAssets,
			balancerPoolAssets:            defaultBalancerAssets,
			uptimeGrowth:                  uptimeHelper.hundredTokensMultiDenom,

			balSharesNotAddedToAccums: true,
			expectedError:             types.BalancerRecordNotFoundError{ClPoolId: 1, BalancerPoolId: 2, UptimeIndex: uint64(0)},
		},
		"insufficient pool balance for balancer distribution": {
			// 100 existing shares and 100 shares added from balancer
			existingConcentratedLiquidity: defaultConcentratedAssets,
			balancerPoolAssets:            defaultBalancerAssets,
			uptimeGrowth:                  uptimeHelper.hundredTokensMultiDenom,

			insufficientPoolBalance: true,
			expectedError:           errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "%s is smaller than %s", sdk.NewCoin("bar", sdk.ZeroInt()), sdk.NewCoin("bar", sdk.NewInt(57000))),
		},
	}
	s.runMultipleAuthorizedUptimes(func() {
		for name, tc := range tests {
			s.Run(name, func() {
				// --- Setup ---

				// Set up CL pool with appropriate liquidity
				s.SetupTest()
				clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], tc.existingConcentratedLiquidity[0].Denom, tc.existingConcentratedLiquidity[1].Denom, DefaultTickSpacing, sdk.ZeroDec())
				clPoolId := clPool.GetId()

				// Set up an existing full range position.
				// Note that the second return value here is the position ID, not an error.
				initialLiquidity, _ := s.SetupPosition(clPoolId, s.TestAccs[0], tc.existingConcentratedLiquidity, DefaultMinTick, DefaultMaxTick, false)

				// Create and bond shares for balancer pool to be linked with CL pool in happy path cases
				balancerPoolId := s.setupBalancerPoolWithFractionLocked(tc.balancerPoolAssets, sdk.OneDec())

				// Invalidate pool IDs if needed for error cases
				if tc.balancerPoolDoesNotExist {
					balancerPoolId = invalidPoolId
				}
				if tc.concentratedPoolDoesNotExist {
					clPoolId = invalidPoolId + 1
				}

				// Link the balancer and CL pools
				s.App.GAMMKeeper.OverwriteMigrationRecordsAndRedirectDistrRecords(s.Ctx,
					gammmigration.MigrationRecords{
						BalancerToConcentratedPoolLinks: []gammmigration.BalancerToConcentratedPoolLink{
							{BalancerPoolId: balancerPoolId, ClPoolId: clPoolId},
						},
					})

				// Add balancer shares to CL accumulatores
				addedLiquidity := sdk.ZeroDec()
				if !tc.balSharesNotAddedToAccums {
					// Get uptime accums for the cl pool.
					uptimeAccums, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, clPool.GetId())
					s.Require().NoError(err)

					addedBalPool, qualifiedShares, err := s.App.ConcentratedLiquidityKeeper.PrepareBalancerPoolAsFullRange(s.Ctx, clPool.GetId(), uptimeAccums)
					addedLiquidity = addedLiquidity.Add(qualifiedShares)

					// If a valid link exists, ensure no error and sanity check the output pool ID
					if !tc.concentratedPoolDoesNotExist && !tc.balancerPoolDoesNotExist {
						s.Require().NoError(err)
						s.Require().Equal(addedBalPool, balancerPoolId)
					}
				}

				// Emit incentives to the uptime accumulators
				for _, growth := range tc.uptimeGrowth {
					decEmissions := growth.MulDec(initialLiquidity.Add(addedLiquidity))
					normalizedEmissions := sdk.NormalizeCoins(decEmissions)

					if !tc.insufficientPoolBalance {
						s.FundAcc(clPool.GetIncentivesAddress(), normalizedEmissions)
					}
				}
				err := addToUptimeAccums(s.Ctx, clPool.GetId(), s.App.ConcentratedLiquidityKeeper, tc.uptimeGrowth)
				s.Require().NoError(err)

				// --- System under test ---

				// Get uptime accums for the cl pool.
				uptimeAccums, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, clPool.GetId())
				s.Require().NoError(err)

				amountClaimed, err := s.App.ConcentratedLiquidityKeeper.ClaimAndResetFullRangeBalancerPool(s.Ctx, clPoolId, balancerPoolId, uptimeAccums)

				// --- Assertions ---

				if tc.expectedError != nil {
					s.Require().ErrorContains(err, tc.expectedError.Error())
					s.Require().Equal(sdk.Coins{}, amountClaimed)

					clPoolUptimeAccumulatorsFromState, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, clPool.GetId())
					s.Require().NoError(err)

					s.Require().True(len(clPoolUptimeAccumulatorsFromState) > 0)
					for uptimeIdx, uptimeAccum := range clPoolUptimeAccumulatorsFromState {
						currAccumShares := uptimeAccum.GetTotalShares()

						// Since reversions for errors are done at a higher level of abstraction,
						// we have to assume that any state updates that happened prior to the error
						// persist for the sake of these unit tests. Thus, balancer full range shares
						// are technically cleared even though in production this process would have been
						// reverted.
						if tc.insufficientPoolBalance {
							addedLiquidity = sdk.ZeroDec()
						}

						expectedLiquidity := initialLiquidity.Add(addedLiquidity)

						// Ensure accum shares remain unchanged after error
						s.Require().Equal(expectedLiquidity, currAccumShares)

						// Also validate uptime accumulators passed in as parameter.
						currAccumShares = uptimeAccums[uptimeIdx].GetTotalShares()
						s.Require().Equal(expectedLiquidity, currAccumShares)
					}

					// If gauge exists, ensure it remains empty after error
					if !tc.balancerPoolDoesNotExist {
						gaugeId, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, balancerPoolId, longestLockableDuration)
						s.Require().NoError(err)

						gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeId)
						s.Require().NoError(err)

						s.Require().Equal(sdk.Coins(nil), gauge.Coins)
					}

					// Ensure amount claimed is zero after error
					s.Require().Equal(sdk.Coins{}, amountClaimed)

					// Pool liquidity should remain unchanged
					updatedClPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPool.GetId())
					s.Require().NoError(err)
					s.Require().Equal(initialLiquidity, updatedClPool.GetLiquidity())

					return
				}

				s.Require().NoError(err)

				clPoolUptimeAccumulators, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, clPool.GetId())
				s.Require().NoError(err)

				s.Require().True(len(clPoolUptimeAccumulators) > 0)
				for uptimeIndex, uptimeAccum := range clPoolUptimeAccumulators {
					currAccumShares := uptimeAccum.GetTotalShares()

					// Ensure each accum has been cleared of the balancer full range shares
					balancerPositionName := string(types.KeyBalancerFullRange(clPoolId, balancerPoolId, uint64(uptimeIndex)))
					fullRangeRecord, err := uptimeAccum.GetPosition(balancerPositionName)
					s.Require().Error(err)
					s.Require().Equal(accum.Record{}, fullRangeRecord)

					// Ensure the full range shares were removed from accum total
					s.Require().Equal(initialLiquidity, currAccumShares)
				}

				// Get balancer gauge corresponding to the longest lockable duration, as this is the one we would be distributing to
				gaugeId, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, balancerPoolId, longestLockableDuration)
				s.Require().NoError(err)
				gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeId)
				s.Require().NoError(err)

				// Since balancer position bonding is a stronger constraint than CL charging, we never need to forfeit incentives
				largestSupportedUptime := s.clk.GetLargestSupportedUptimeDuration(s.Ctx)

				// Calculate the number of tokens we expect to see in the balancer gauge
				expectedTokensInGauge := expectedIncentivesFromUptimeGrowth(tc.uptimeGrowth, addedLiquidity, largestSupportedUptime, defaultMultiplier)

				// Ensure gauge coins and amountClaimed are correct
				s.Require().Equal(expectedTokensInGauge.String(), gauge.Coins.String())
				s.Require().Equal(expectedTokensInGauge.String(), amountClaimed.String())

				// Pool liquidity should remain unchanged
				updatedClPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPool.GetId())
				s.Require().NoError(err)
				s.Require().Equal(initialLiquidity, updatedClPool.GetLiquidity())
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

				clParams := s.clk.GetParams(s.Ctx)
				clParams.AuthorizedUptimes = tc.preSetAuthorizedParams
				s.clk.SetParams(s.Ctx, clParams)

				actualAuthorized := s.clk.GetLargestAuthorizedUptimeDuration(s.Ctx)
				actualSupported := s.clk.GetLargestSupportedUptimeDuration(s.Ctx)

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
	testAccumulator, err := s.clk.GetSpreadRewardAccumulator(s.Ctx, pool.GetId())
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
			growthOutside:            defaultGlobalRewardGrowth.QuoDec(sdk.NewDec(2)),
			numShares:                1,
			expectedUnclaimedRewards: defaultGlobalRewardGrowth.QuoDec(sdk.NewDec(2)),
		},
		"multiple shares, partial growth outside": {
			growthOutside:            defaultGlobalRewardGrowth.QuoDec(sdk.NewDec(2)),
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

				err := testAccumulator.NewPosition(oldPos, sdk.NewDec(tc.numShares), nil)
				s.Require().NoError(err)

				// 2 shares is chosen arbitrarily. It is not relevant for this test.
				err = testAccumulator.NewPosition(newPos, sdk.NewDec(2), nil)
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
	foo100 := sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(100)))
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
				{UptimeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin("fooa", sdk.NewInt(100)))},
				{UptimeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin("foob", sdk.NewInt(100)))},
			},
			expectedOutput: []sdk.DecCoins{
				foo100,
				sdk.NewDecCoins(sdk.NewDecCoin("fooa", sdk.NewInt(100))),
				sdk.NewDecCoins(sdk.NewDecCoin("foob", sdk.NewInt(100))),
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
						RemainingCoin: sdk.NewDecCoinFromDec(USDC, sdk.NewDec(1000)),
						EmissionRate:  sdk.NewDec(1), // 1 per second
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
