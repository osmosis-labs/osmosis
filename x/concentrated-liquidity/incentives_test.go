package concentrated_liquidity_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var (
	defaultPoolId     = uint64(1)
	defaultMultiplier = sdk.OneInt()

	testAddressOne   = sdk.MustAccAddressFromBech32("osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks")
	testAddressTwo   = sdk.MustAccAddressFromBech32("osmo18s5lynnmx37hq4wlrw9gdn68sg2uxp5rgk26vv")
	testAddressThree = sdk.MustAccAddressFromBech32("osmo1qwexv7c6sm95lwhzn9027vyu2ccneaqad4w8ka")
	testAddressFour  = sdk.MustAccAddressFromBech32("osmo14hcxlnwlqtq75ttaxf674vk6mafspg8xwgnn53")

	testAccumOne = "testAccumOne"

	// Note: lexicographic order is denomFour, denomOne, denomThree, denomTwo
	testDenomOne   = "denomOne"
	testDenomTwo   = "denomTwo"
	testDenomThree = "denomThree"
	testDenomFour  = "denomFour"

	defaultIncentiveAmount = sdk.NewDec(2 << 60)

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
		PoolId:               validPoolId,
		IncentiveDenom:       testDenomOne,
		IncentiveCreatorAddr: testAddressOne.String(),
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingAmount: defaultIncentiveAmount,
			EmissionRate:    testEmissionOne,
			StartTime:       defaultStartTime,
		},
		MinUptime: testUptimeOne,
	}

	incentiveRecordTwo = types.IncentiveRecord{
		PoolId:               validPoolId,
		IncentiveDenom:       testDenomTwo,
		IncentiveCreatorAddr: testAddressTwo.String(),
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingAmount: defaultIncentiveAmount,
			EmissionRate:    testEmissionTwo,
			StartTime:       defaultStartTime,
		},
		MinUptime: testUptimeTwo,
	}

	incentiveRecordThree = types.IncentiveRecord{
		PoolId:               validPoolId,
		IncentiveDenom:       testDenomThree,
		IncentiveCreatorAddr: testAddressThree.String(),
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingAmount: defaultIncentiveAmount,
			EmissionRate:    testEmissionThree,
			StartTime:       defaultStartTime,
		},
		MinUptime: testUptimeThree,
	}

	incentiveRecordFour = types.IncentiveRecord{
		PoolId:               validPoolId,
		IncentiveDenom:       testDenomFour,
		IncentiveCreatorAddr: testAddressFour.String(),
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingAmount: defaultIncentiveAmount,
			EmissionRate:    testEmissionFour,
			StartTime:       defaultStartTime,
		},
		MinUptime: testUptimeFour,
	}

	emptyIncentiveRecord = types.IncentiveRecord{
		PoolId:               validPoolId,
		IncentiveDenom:       "emptyDenom",
		IncentiveCreatorAddr: testAddressFour.String(),
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingAmount: sdk.ZeroDec(),
			EmissionRate:    testEmissionFour,
			StartTime:       defaultStartTime,
		},
		MinUptime: testUptimeFour,
	}

	testQualifyingDepositsOne   = sdk.NewInt(50)
	testQualifyingDepositsTwo   = sdk.NewInt(100)
	testQualifyingDepositsThree = sdk.NewInt(399)
)

type ExpectedUptimes struct {
	emptyExpectedAccumValues     []sdk.DecCoins
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
	totalRewards := sdk.DecCoins(nil)
	for uptimeIndex, uptimeGrowth := range uptimeGrowths {
		if timeInPool >= types.SupportedUptimes[uptimeIndex] {
			totalRewards = totalRewards.Add(uptimeGrowth...)
		}
	}

	// Calculate position's pro-rata share by multiplying growth by position's shares
	positionRewards := sdk.Coins(nil)
	for _, rewardToken := range totalRewards {
		positionRewards = positionRewards.Add(sdk.NewCoin(rewardToken.Denom, rewardToken.Amount.Mul(positionShares).TruncateInt().Mul(multiplier)))
	}

	return positionRewards
}

func chargeIncentive(incentiveRecord types.IncentiveRecord, timeElapsed time.Duration) types.IncentiveRecord {
	incentivesEmitted := incentiveRecord.IncentiveRecordBody.EmissionRate.Mul(sdk.NewDec(int64(timeElapsed)).Quo(sdk.MustNewDecFromStr("1000000000")))
	incentiveRecord.IncentiveRecordBody.RemainingAmount = incentiveRecord.IncentiveRecordBody.RemainingAmount.Sub(incentivesEmitted)

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

// addDecCoinsArray adds the contents of the second param from the first (decCoinsArrayA + decCoinsArrayB)
// Note that this takes in two _arrays_ of DecCoins, meaning that each term itself is of type DecCoins (i.e. an array of DecCoin).
func addDecCoinsArray(decCoinsArrayA []sdk.DecCoins, decCoinsArrayB []sdk.DecCoins) ([]sdk.DecCoins, error) {
	if len(decCoinsArrayA) != len(decCoinsArrayB) {
		return []sdk.DecCoins{}, fmt.Errorf("DecCoin arrays must be of equal length to be added")
	}

	finalDecCoinArray := []sdk.DecCoins{}
	for i := range decCoinsArrayA {
		finalDecCoinArray = append(finalDecCoinArray, decCoinsArrayA[i].Add(decCoinsArrayB[i]...))
	}

	return finalDecCoinArray, nil
}

func createIncentiveRecord(incentiveDenom string, remainingAmt, emissionRate sdk.Dec, startTime time.Time, minUpTime time.Duration) types.IncentiveRecord {
	return types.IncentiveRecord{
		IncentiveDenom: incentiveDenom,
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingAmount: remainingAmt,
			EmissionRate:    emissionRate,
			StartTime:       startTime,
		},
		MinUptime: minUpTime,
	}
}

func withDenom(record types.IncentiveRecord, denom string) types.IncentiveRecord {
	record.IncentiveDenom = denom

	return record
}

func withAmount(record types.IncentiveRecord, amount sdk.Dec) types.IncentiveRecord {
	record.IncentiveRecordBody.RemainingAmount = amount

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
			expectedAccumName: string(types.KeyUptimeAccumulator(1, 0)),
		},
		"pool id 1, uptime id 999": {
			poolId:            defaultPoolId,
			uptimeIndex:       uint64(999),
			expectedAccumName: string(types.KeyUptimeAccumulator(1, 999)),
		},
		"pool id 999, uptime id 1": {
			poolId:            uint64(999),
			uptimeIndex:       uint64(1),
			expectedAccumName: string(types.KeyUptimeAccumulator(999, 1)),
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()

			// system under test
			accumName := types.KeyUptimeAccumulator(tc.poolId, tc.uptimeIndex)
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
	incentiveRecordOneWithDifferentStartTime := withStartTime(incentiveRecordOne, incentiveRecordOne.IncentiveRecordBody.StartTime.Add(10))
	incentiveRecordOneWithDifferentMinUpTime := withMinUptime(incentiveRecordOne, testUptimeTwo)
	incentiveRecordOneWithDifferentDenom := withDenom(incentiveRecordOne, testDenomTwo)

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
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveDenom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{chargeIncentive(incentiveRecordOne, time.Hour)},
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
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveDenom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{
				// We only charge the first incentive record since the second wasn't affected
				chargeIncentive(incentiveRecordOne, time.Hour),
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
				sdk.NewDecCoinFromDec(incentiveRecordOne.IncentiveDenom, incentiveRecordOne.IncentiveRecordBody.RemainingAmount.QuoTruncate(sdk.NewDec(123))),
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
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveDenom, incentiveRecordOne.IncentiveRecordBody.EmissionRate.Add(incentiveRecordOneWithDifferentStartTime.IncentiveRecordBody.EmissionRate), time.Hour, sdk.NewDec(100)), // since we have 2 records with same denom, the rate of emission went up x2
			),
			expectedIncentiveRecords: []types.IncentiveRecord{
				// We only going to charge both incentive records
				chargeIncentive(incentiveRecordOne, time.Hour),
				chargeIncentive(incentiveRecordOneWithDifferentStartTime, time.Hour),
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
				expectedIncentivesFromRate(incentiveRecordOneWithDifferentStartTime.IncentiveDenom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
				expectedIncentivesFromRate(incentiveRecordOneWithDifferentDenom.IncentiveDenom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{
				// We charge both incentive record here because both minUpTime has been hit
				chargeIncentive(incentiveRecordOneWithDifferentStartTime, time.Hour),
				chargeIncentive(incentiveRecordOneWithDifferentDenom, time.Hour),
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
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveDenom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{
				// We only charge the first incentive record because the second minUpTime hasn't been hit yet
				chargeIncentive(incentiveRecordOne, time.Hour),
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
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveDenom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
				expectedIncentivesFromRate(incentiveRecordOneWithDifferentDenom.IncentiveDenom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
			},
			expectedIncentiveRecords: []types.IncentiveRecord{
				// We charge both incentive record here because both minUpTime has been hit
				chargeIncentive(incentiveRecordOne, time.Hour),
				chargeIncentive(incentiveRecordOneWithDifferentDenom, time.Hour),
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
				expectedIncentivesFromRate(incentiveRecordOne.IncentiveDenom, incentiveRecordOne.IncentiveRecordBody.EmissionRate.Add(incentiveRecordOneWithDifferentStartTime.IncentiveRecordBody.EmissionRate), time.Hour, sdk.NewDec(100)),
				expectedIncentivesFromRate(incentiveRecordOneWithDifferentDenom.IncentiveDenom, incentiveRecordOne.IncentiveRecordBody.EmissionRate, time.Hour, sdk.NewDec(100)),
			),
			expectedIncentiveRecords: []types.IncentiveRecord{
				// We only charge the first three incentive record because the fourth minUpTime hasn't been hit yet
				chargeIncentive(incentiveRecordOne, time.Hour),
				chargeIncentive(incentiveRecordOneWithDifferentStartTime, time.Hour),
				chargeIncentive(incentiveRecordOneWithDifferentDenom, time.Hour),
				incentiveRecordOneWithDifferentMinUpTime, // this uptime hasn't hit yet so we do not have to charge incentive
			},
			expectedPass: true,
		},
	}

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
					err := s.App.ConcentratedLiquidityKeeper.SetMultipleIncentiveRecords(s.Ctx, updatedPoolRecords)
					s.Require().NoError(err)

					updatedRecordsInState, err := s.App.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(s.Ctx, tc.poolId)
					s.Require().NoError(err)

					s.Require().Equal(0, len(updatedRecordsInState))
				}
			} else {
				s.Require().Error(err)
			}
		})
	}
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
		accumUptime          time.Duration
		qualifyingLiquidity  sdk.Dec
		timeElapsed          time.Duration
		poolIncentiveRecords []types.IncentiveRecord

		expectedResult           sdk.DecCoins
		expectedUptimeDeltas     []sdk.DecCoins
		expectedIncentiveRecords []types.IncentiveRecord
		expectedPass             bool
	}
	tests := map[string]updateAccumToNow{
		"one incentive record": {
			poolId:               defaultPoolId,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne},

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We deduct incentives from the record for the period it emitted incentives
				chargeIncentive(incentiveRecordOne, defaultTestUptime+time.Hour),
			},
			expectedPass: true,
		},
		"two incentive records, each with qualifying liquidity": {
			poolId:               defaultPoolId,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo},

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We deduct incentives from both records since there are positions for each
				chargeIncentive(incentiveRecordOne, defaultTestUptime+time.Hour),
				chargeIncentive(incentiveRecordTwo, defaultTestUptime+time.Hour),
			},
			expectedPass: true,
		},
		"three incentive records, each with qualifying liquidity": {
			poolId:               defaultPoolId,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo, incentiveRecordThree},

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We deduct incentives from each record since there are positions for all three
				// Note that records are in ascending order by uptime index
				chargeIncentive(incentiveRecordOne, defaultTestUptime+time.Hour),
				chargeIncentive(incentiveRecordTwo, defaultTestUptime+time.Hour),
				chargeIncentive(incentiveRecordThree, defaultTestUptime+time.Hour),
			},
			expectedPass: true,
		},
		"two incentive records, only one with qualifying liquidity": {
			poolId:               defaultPoolId,
			timeElapsed:          time.Hour,
			poolIncentiveRecords: []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo, incentiveRecordThree, incentiveRecordFour},

			expectedIncentiveRecords: []types.IncentiveRecord{
				// We only deduct from the first three incentive records since the last doesn't emit anything
				// Note that records are in ascending order by uptime index
				chargeIncentive(incentiveRecordOne, defaultTestUptime+time.Hour),
				chargeIncentive(incentiveRecordTwo, defaultTestUptime+time.Hour),
				chargeIncentive(incentiveRecordThree, defaultTestUptime+time.Hour),
				// We charge even for uptimes the position has technically not qualified for since its liquidity is on
				// the accumulator.
				chargeIncentive(incentiveRecordFour, defaultTestUptime+time.Hour),
			},
			expectedPass: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			clKeeper := s.App.ConcentratedLiquidityKeeper
			s.Ctx = s.Ctx.WithBlockTime(defaultStartTime)

			// Set up test pool
			clPool := s.PrepareConcentratedPool()

			// Initialize test incentives on the pool
			err := clKeeper.SetMultipleIncentiveRecords(s.Ctx, tc.poolIncentiveRecords)
			s.Require().NoError(err)

			// Get initial uptime accum values for comparison
			initUptimeAccumValues, err := clKeeper.GetUptimeAccumulatorValues(s.Ctx, tc.poolId)
			s.Require().NoError(err)

			// Add qualifying and non-qualifying liquidity to the pool
			s.FundAcc(testAddressOne, sdk.NewCoins(sdk.NewCoin(clPool.GetToken0(), testQualifyingDepositsOne), sdk.NewCoin(clPool.GetToken1(), testQualifyingDepositsOne)))
			_, _, _, qualifyingLiquidity, _, err := clKeeper.CreatePosition(s.Ctx, tc.poolId, testAddressOne, testQualifyingDepositsOne, testQualifyingDepositsOne, sdk.ZeroInt(), sdk.ZeroInt(), clPool.GetCurrentTick().Int64()-1, clPool.GetCurrentTick().Int64()+1)
			s.Require().NoError(err)

			// Let enough time elapse to qualify the position for the first three supported uptimes
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(defaultTestUptime))

			// Let `timeElapsed` time pass to test incentive distribution
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(tc.timeElapsed))

			// System under test
			err = clKeeper.UpdateUptimeAccumulatorsToNow(s.Ctx, tc.poolId)

			if tc.expectedPass {
				s.Require().NoError(err)

				// Get updated pool for testing purposes
				clPool, err := clKeeper.GetPoolById(s.Ctx, tc.poolId)
				s.Require().NoError(err)

				// Get new uptime accum values for comparison
				newUptimeAccumValues, err := clKeeper.GetUptimeAccumulatorValues(s.Ctx, tc.poolId)
				s.Require().NoError(err)

				// Calculate expected uptime deltas using qualifying liquidity deltas
				expectedUptimeDeltas := []sdk.DecCoins{}
				for uptimeIndex := range newUptimeAccumValues {
					// Calculate expected incentives for the current uptime by emitting incentives from
					// all incentive records for the it
					curUptimeAccruedIncentives := cl.EmptyCoins
					for _, poolRecord := range tc.poolIncentiveRecords {
						if poolRecord.MinUptime == types.SupportedUptimes[uptimeIndex] {
							// We set the expected accrued incentives based on the total time that has elapsed since position creation
							curUptimeAccruedIncentives = curUptimeAccruedIncentives.Add(sdk.NewDecCoins(expectedIncentivesFromRate(poolRecord.IncentiveDenom, poolRecord.IncentiveRecordBody.EmissionRate, defaultTestUptime+tc.timeElapsed, qualifyingLiquidity))...)
						}
					}
					expectedUptimeDeltas = append(expectedUptimeDeltas, curUptimeAccruedIncentives)
				}

				// Ensure that each accumulator value changes by the correct amount
				for uptimeIndex := range newUptimeAccumValues {
					uptimeDelta := newUptimeAccumValues[uptimeIndex].Sub(initUptimeAccumValues[uptimeIndex])
					s.Require().Equal(expectedUptimeDeltas[uptimeIndex], uptimeDelta)
				}

				// Ensure that LastLiquidityUpdate field is updated for pool
				s.Require().Equal(s.Ctx.BlockTime(), clPool.GetLastLiquidityUpdate())
				// Ensure that pool's IncentiveRecords are updated to reflect emitted incentives
				updatedIncentiveRecords, err := clKeeper.GetAllIncentiveRecordsForPool(s.Ctx, tc.poolId)
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedIncentiveRecords, updatedIncentiveRecords)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// Note: we test that incentive records are properly deducted by emissions in `TestUpdateUptimeAccumulatorsToNow` above.
// This test aims to cover the behavior of a series of state read/writes relating to incentive records.
func (s *KeeperTestSuite) TestIncentiveRecordsSetAndGet() {
	s.SetupTest()
	clKeeper := s.App.ConcentratedLiquidityKeeper
	s.Ctx = s.Ctx.WithBlockTime(defaultStartTime)
	emptyIncentiveRecords := []types.IncentiveRecord{}

	// Set up test pool
	clPoolOne := s.PrepareConcentratedPool()

	// Set up second pool for reference
	clPoolTwo := s.PrepareConcentratedPool()

	// Ensure both pools start with no incentive records
	poolOneRecords, err := clKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolOne.GetId())
	s.Require().NoError(err)
	s.Require().Equal(emptyIncentiveRecords, poolOneRecords)

	poolTwoRecords, err := clKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolTwo.GetId())
	s.Require().NoError(err)
	s.Require().Equal(emptyIncentiveRecords, poolTwoRecords)

	// Ensure setting and getting a single record works with single Get and GetAll
	clKeeper.SetIncentiveRecord(s.Ctx, incentiveRecordOne)
	poolOneRecord, err := clKeeper.GetIncentiveRecord(s.Ctx, clPoolOne.GetId(), incentiveRecordOne.IncentiveDenom, incentiveRecordOne.MinUptime, sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr))
	s.Require().NoError(err)
	s.Require().Equal(incentiveRecordOne, poolOneRecord)
	allRecordsPoolOne, err := clKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolOne.GetId())
	s.Require().NoError(err)
	s.Require().Equal([]types.IncentiveRecord{incentiveRecordOne}, allRecordsPoolOne)

	// Ensure records for other pool remain unchanged
	poolTwoRecord, err := clKeeper.GetIncentiveRecord(s.Ctx, clPoolTwo.GetId(), incentiveRecordOne.IncentiveDenom, incentiveRecordOne.MinUptime, sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr))
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.IncentiveRecordNotFoundError{PoolId: clPoolTwo.GetId(), IncentiveDenom: incentiveRecordOne.IncentiveDenom, MinUptime: incentiveRecordOne.MinUptime, IncentiveCreatorStr: incentiveRecordOne.IncentiveCreatorAddr})
	s.Require().Equal(types.IncentiveRecord{}, poolTwoRecord)
	allRecordsPoolTwo, err := clKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolTwo.GetId())
	s.Require().NoError(err)
	s.Require().Equal(emptyIncentiveRecords, allRecordsPoolTwo)

	// Ensure directly setting additional records don't overwrite previous ones
	clKeeper.SetIncentiveRecord(s.Ctx, incentiveRecordTwo)
	poolOneRecord, err = clKeeper.GetIncentiveRecord(s.Ctx, clPoolOne.GetId(), incentiveRecordTwo.IncentiveDenom, incentiveRecordTwo.MinUptime, sdk.MustAccAddressFromBech32(incentiveRecordTwo.IncentiveCreatorAddr))
	s.Require().NoError(err)
	s.Require().Equal(incentiveRecordTwo, poolOneRecord)
	allRecordsPoolOne, err = clKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolOne.GetId())
	s.Require().NoError(err)
	s.Require().Equal([]types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo}, allRecordsPoolOne)

	// Ensure setting multiple records through helper functions as expected
	// Note that we also pass in an empty incentive record, which we expect to be cleared out while being set
	err = clKeeper.SetMultipleIncentiveRecords(s.Ctx, []types.IncentiveRecord{incentiveRecordThree, incentiveRecordFour, emptyIncentiveRecord})
	s.Require().NoError(err)

	// Note: we expect the records to be retrieved in lexicographic order by denom and for the empty record to be cleared
	allRecordsPoolOne, err = clKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolOne.GetId())
	s.Require().NoError(err)
	s.Require().Equal([]types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo, incentiveRecordThree, incentiveRecordFour}, allRecordsPoolOne)

	// Finally, we ensure the second pool remains unaffected
	allRecordsPoolTwo, err = clKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolTwo.GetId())
	s.Require().NoError(err)
	s.Require().Equal(emptyIncentiveRecords, allRecordsPoolTwo)
}

func (s *KeeperTestSuite) TestGetInitialUptimeGrowthOutsidesForTick() {
	expectedUptimes := getExpectedUptimes()

	type getInitialUptimeGrowthOutsidesForTick struct {
		poolId                          uint64
		tick                            int64
		expectedUptimeAccumulatorValues []sdk.DecCoins
	}
	tests := map[string]getInitialUptimeGrowthOutsidesForTick{
		"uptime growth for tick <= currentTick": {
			poolId:                          1,
			tick:                            -2,
			expectedUptimeAccumulatorValues: expectedUptimes.hundredTokensMultiDenom,
		},
		"uptime growth for tick > currentTick": {
			poolId:                          2,
			tick:                            1,
			expectedUptimeAccumulatorValues: expectedUptimes.emptyExpectedAccumValues,
		},
	}

	for name, tc := range tests {
		tc := tc

		s.Run(name, func() {
			s.SetupTest()
			clKeeper := s.App.ConcentratedLiquidityKeeper

			// create 2 pools
			s.PrepareConcentratedPool()
			s.PrepareConcentratedPool()

			poolUptimeAccumulators, err := clKeeper.GetUptimeAccumulators(s.Ctx, tc.poolId)
			s.Require().NoError(err)

			for uptimeId, uptimeAccum := range poolUptimeAccumulators {
				uptimeAccum.AddToAccumulator(expectedUptimes.hundredTokensMultiDenom[uptimeId])
			}

			poolUptimeAccumulators, err = clKeeper.GetUptimeAccumulators(s.Ctx, tc.poolId)
			s.Require().NoError(err)

			val, err := clKeeper.GetInitialUptimeGrowthOutsidesForTick(s.Ctx, tc.poolId, tc.tick)
			s.Require().NoError(err)
			s.Require().Equal(val, tc.expectedUptimeAccumulatorValues)
		})
	}
}

func (s *KeeperTestSuite) TestGetUptimeGrowthInsideRange() {
	defaultPoolId := uint64(1)
	defaultInitialLiquidity := sdk.OneDec()
	uptimeHelper := getExpectedUptimes()

	type uptimeGrowthOutsideTest struct {
		poolSetup bool

		lowerTick                    int64
		upperTick                    int64
		currentTick                  int64
		lowerTickUptimeGrowthOutside []sdk.DecCoins
		upperTickUptimeGrowthOutside []sdk.DecCoins
		globalUptimeGrowth           []sdk.DecCoins

		expectedUptimeGrowthInside []sdk.DecCoins
		invalidTick                bool
		expectedError              bool
	}

	tests := map[string]uptimeGrowthOutsideTest{
		// current tick above range

		"current tick > upper tick, nonzero uptime growth inside": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  2,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is above range, we expect upper - lower
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			expectedError:              false,
		},
		"current tick > upper tick, nonzero uptime growth inside (wider range)": {
			poolSetup:                    true,
			lowerTick:                    12444,
			upperTick:                    15013,
			currentTick:                  50320,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is above range, we expect upper - lower
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			expectedError:              false,
		},
		"current tick > upper tick, zero uptime growth inside (nonempty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  2,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is above range, we expect upper - lower
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:              false,
		},
		"current tick > upper tick, zero uptime growth inside (empty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  2,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since current tick is above range, we expect upper - lower
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:              false,
		},
		"current tick > upper tick, zero uptime growth inside with extraneous uptime growth": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  2,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.twoHundredTokensMultiDenom,

			// Since current tick is above range, we expect upper - lower
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:              false,
		},

		// current tick within range

		"upper tick > current tick > lower tick, nonzero uptime growth inside": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    2,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is within range, we expect global - upper - lower
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			expectedError:              false,
		},
		"upper tick > current tick > lower tick, nonzero uptime growth inside (wider range)": {
			poolSetup:                    true,
			lowerTick:                    -19753,
			upperTick:                    8921,
			currentTick:                  -97,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is within range, we expect global - upper - lower
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			expectedError:              false,
		},
		"upper tick > current tick > lower tick, zero uptime growth inside (nonempty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    2,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.twoHundredTokensMultiDenom,

			// Since current tick is within range, we expect global - upper - lower
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:              false,
		},
		"upper tick > current tick > lower tick, zero uptime growth inside (empty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    2,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since current tick is within range, we expect global - upper - lower
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:              false,
		},

		// current tick below range

		"current tick < lower tick, nonzero uptime growth inside": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  -1,
			lowerTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is below range, we expect lower - upper
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			expectedError:              false,
		},
		"current tick < lower tick, nonzero uptime growth inside (wider range)": {
			poolSetup:                    true,
			lowerTick:                    328,
			upperTick:                    726,
			currentTick:                  189,
			lowerTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is below range, we expect lower - upper
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			expectedError:              false,
		},
		"current tick < lower tick, zero uptime growth inside (nonempty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  -1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is below range, we expect lower - upper
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:              false,
		},
		"current tick < lower tick, zero uptime growth inside (empty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  -1,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since current tick is below range, we expect lower - upper
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:              false,
		},

		// current tick on range boundary

		"current tick = lower tick, nonzero uptime growth inside": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  0,
			lowerTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.fourHundredTokensMultiDenom,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being within the range (global - upper - lower)
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			expectedError:              false,
		},
		"current tick = lower tick, zero uptime growth inside (nonempty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  0,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.twoHundredTokensMultiDenom,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being within the range (global - upper - lower)
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:              false,
		},
		"current tick = lower tick, zero uptime growth inside (empty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  0,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being within the range (global - upper - lower)
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:              false,
		},
		"current tick = upper tick, nonzero uptime growth inside": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.fourHundredTokensMultiDenom,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being above the range (upper - lower)
			expectedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			expectedError:              false,
		},
		"current tick = upper tick, zero uptime growth inside (nonempty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.twoHundredTokensMultiDenom,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being above the range (upper - lower)
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:              false,
		},
		"current tick = upper tick, zero uptime growth inside (empty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being above the range (upper - lower)
			expectedUptimeGrowthInside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:              false,
		},

		// error catching

		"error: pool has not been setup": {
			poolSetup:     false,
			expectedError: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			// if pool set up true, set up default pool
			var pool types.ConcentratedPoolExtension
			if tc.poolSetup {
				pool = s.PrepareConcentratedPool()
				currentTick := pool.GetCurrentTick().Int64()

				// Update global uptime accums
				addToUptimeAccums(s.Ctx, pool.GetId(), s.App.ConcentratedLiquidityKeeper, tc.globalUptimeGrowth)

				// Update tick-level uptime trackers
				s.initializeTick(s.Ctx, currentTick, tc.lowerTick, defaultInitialLiquidity, cl.EmptyCoins, wrapUptimeTrackers(tc.lowerTickUptimeGrowthOutside), true)
				s.initializeTick(s.Ctx, currentTick, tc.upperTick, defaultInitialLiquidity, cl.EmptyCoins, wrapUptimeTrackers(tc.upperTickUptimeGrowthOutside), false)
				pool.SetCurrentTick(sdk.NewInt(tc.currentTick))
				s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, pool)
			}

			// system under test
			uptimeGrowthInside, err := s.App.ConcentratedLiquidityKeeper.GetUptimeGrowthInsideRange(s.Ctx, defaultPoolId, tc.lowerTick, tc.upperTick)
			if tc.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				// check if returned uptime growth inside has correct value
				s.Require().Equal(tc.expectedUptimeGrowthInside, uptimeGrowthInside)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetUptimeGrowthOutsideRange() {
	defaultPoolId := uint64(1)
	defaultInitialLiquidity := sdk.OneDec()
	uptimeHelper := getExpectedUptimes()

	type uptimeGrowthOutsideTest struct {
		poolSetup bool

		lowerTick                    int64
		upperTick                    int64
		currentTick                  int64
		lowerTickUptimeGrowthOutside []sdk.DecCoins
		upperTickUptimeGrowthOutside []sdk.DecCoins
		globalUptimeGrowth           []sdk.DecCoins

		expectedUptimeGrowthOutside []sdk.DecCoins
		invalidTick                 bool
		expectedError               bool
	}

	tests := map[string]uptimeGrowthOutsideTest{
		// current tick above range

		"current tick > upper tick, nonzero uptime growth inside": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  2,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is above range, we expect global - (upper - lower)
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			expectedError:               false,
		},
		"current tick > upper tick, nonzero uptime growth inside (wider range)": {
			poolSetup:                    true,
			lowerTick:                    12444,
			upperTick:                    15013,
			currentTick:                  50320,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is above range, we expect global - (upper - lower)
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			expectedError:               false,
		},
		"current tick > upper tick, zero uptime growth inside (nonempty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  2,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.threeHundredTokensMultiDenom,
			expectedError:               false,
		},
		"current tick > upper tick, zero uptime growth inside (empty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  2,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:               false,
		},
		"current tick > upper tick, zero uptime growth inside with extraneous uptime growth": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  2,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.twoHundredTokensMultiDenom,

			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			expectedError:               false,
		},

		// current tick within range

		"upper tick > current tick > lower tick, nonzero uptime growth inside": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    2,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is within range, we expect global - (global - upper - lower)
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			expectedError:               false,
		},
		"upper tick > current tick > lower tick, nonzero uptime growth inside (wider range)": {
			poolSetup:                    true,
			lowerTick:                    -19753,
			upperTick:                    8921,
			currentTick:                  -97,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.twoHundredTokensMultiDenom,

			// Since current tick is within range, we expect global - (global - upper - lower)
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			expectedError:               false,
		},
		"upper tick > current tick > lower tick, zero uptime growth inside (nonempty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    2,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.twoHundredTokensMultiDenom,

			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			expectedError:               false,
		},
		"upper tick > current tick > lower tick, zero uptime growth inside (empty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    2,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:               false,
		},

		// current tick below range

		"current tick < lower tick, nonzero uptime growth inside": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  -1,
			lowerTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is below range, we expect global - (lower - upper)
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			expectedError:               false,
		},
		"current tick < lower tick, nonzero uptime growth inside (wider range)": {
			poolSetup:                    true,
			lowerTick:                    328,
			upperTick:                    726,
			currentTick:                  189,
			lowerTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since current tick is below range, we expect global - (lower - upper)
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			expectedError:               false,
		},
		"current tick < lower tick, zero uptime growth inside (nonempty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  -1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.threeHundredTokensMultiDenom,

			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.threeHundredTokensMultiDenom,
			expectedError:               false,
		},
		"current tick < lower tick, zero uptime growth inside (empty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  -1,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:               false,
		},

		// current tick on range boundary

		"current tick = lower tick, nonzero uptime growth inside": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  0,
			lowerTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.fourHundredTokensMultiDenom,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being within the range (global - (global - upper - lower))
			expectedUptimeGrowthOutside: uptimeHelper.threeHundredTokensMultiDenom,
			expectedError:               false,
		},
		"current tick = lower tick, zero uptime growth inside (nonempty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  0,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.twoHundredTokensMultiDenom,

			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			expectedError:               false,
		},
		"current tick = lower tick, zero uptime growth inside (empty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  0,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:               false,
		},
		"current tick = upper tick, nonzero uptime growth inside": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.fourHundredTokensMultiDenom,

			// Since we treat the range as [lower, upper) (i.e. inclusive of lower tick, exclusive of upper),
			// this case is equivalent to the current tick being above the range (global - (upper - lower))
			expectedUptimeGrowthOutside: uptimeHelper.threeHundredTokensMultiDenom,
			expectedError:               false,
		},
		"current tick = upper tick, zero uptime growth inside (nonempty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			upperTickUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			globalUptimeGrowth:           uptimeHelper.twoHundredTokensMultiDenom,

			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.twoHundredTokensMultiDenom,
			expectedError:               false,
		},
		"current tick = upper tick, zero uptime growth inside (empty trackers)": {
			poolSetup:                    true,
			lowerTick:                    0,
			upperTick:                    1,
			currentTick:                  1,
			lowerTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			upperTickUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			globalUptimeGrowth:           uptimeHelper.emptyExpectedAccumValues,

			// Since the range is empty, we expect growth outside to be equal to global
			expectedUptimeGrowthOutside: uptimeHelper.emptyExpectedAccumValues,
			expectedError:               false,
		},

		// error catching

		"error: pool has not been setup": {
			poolSetup:     false,
			expectedError: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			// if pool set up true, set up default pool
			var pool types.ConcentratedPoolExtension
			if tc.poolSetup {
				pool = s.PrepareConcentratedPool()
				currentTick := pool.GetCurrentTick().Int64()

				// Update global uptime accums
				addToUptimeAccums(s.Ctx, pool.GetId(), s.App.ConcentratedLiquidityKeeper, tc.globalUptimeGrowth)

				// Update tick-level uptime trackers
				s.initializeTick(s.Ctx, currentTick, tc.lowerTick, defaultInitialLiquidity, cl.EmptyCoins, wrapUptimeTrackers(tc.lowerTickUptimeGrowthOutside), true)
				s.initializeTick(s.Ctx, currentTick, tc.upperTick, defaultInitialLiquidity, cl.EmptyCoins, wrapUptimeTrackers(tc.upperTickUptimeGrowthOutside), false)
				pool.SetCurrentTick(sdk.NewInt(tc.currentTick))
				s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, pool)
			}

			// system under test
			uptimeGrowthOutside, err := s.App.ConcentratedLiquidityKeeper.GetUptimeGrowthOutsideRange(s.Ctx, defaultPoolId, tc.lowerTick, tc.upperTick)
			if tc.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				// check if returned uptime growth inside has correct value
				s.Require().Equal(tc.expectedUptimeGrowthOutside, uptimeGrowthOutside)
			}
		})
	}
}

func (s *KeeperTestSuite) TestInitPositionUptime() {
	uptimeHelper := getExpectedUptimes()
	type tick struct {
		tickIndex      int64
		uptimeTrackers []model.UptimeTracker
	}

	tests := []struct {
		name              string
		positionLiquidity sdk.Dec

		lowerTick               tick
		upperTick               tick
		positionId              uint64
		currentTickIndex        sdk.Int
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

		{
			name:              "(lower < curr < upper) nonzero uptime trackers",
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
			currentTickIndex:         sdk.ZeroInt(),
			globalUptimeAccumValues:  uptimeHelper.threeHundredTokensMultiDenom,
			expectedInitAccumValue:   uptimeHelper.hundredTokensMultiDenom,
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
		{
			name:              "(lower < upper < curr) nonzero uptime trackers",
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
			currentTickIndex:         sdk.NewInt(51),
			globalUptimeAccumValues:  uptimeHelper.fourHundredTokensMultiDenom,
			expectedInitAccumValue:   uptimeHelper.twoHundredTokensMultiDenom,
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
		{
			name:              "(curr < lower < upper) nonzero uptime trackers",
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
			currentTickIndex:         sdk.NewInt(-51),
			globalUptimeAccumValues:  uptimeHelper.fourHundredTokensMultiDenom,
			expectedInitAccumValue:   uptimeHelper.twoHundredTokensMultiDenom,
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
		{
			name:              "(lower < curr < upper) nonzero and variable uptime trackers",
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
			currentTickIndex: sdk.ZeroInt(),

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
			},
			// Equal to 100 of foo and bar in each uptime tracker (UGI)
			expectedInitAccumValue:   uptimeHelper.hundredTokensMultiDenom,
			expectedUnclaimedRewards: uptimeHelper.emptyExpectedAccumValues,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// --- Setup ---

			// Init suite for each test.
			s.Setup()

			// Set blocktime to fixed UTC value for consistency
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			clPool := s.PrepareConcentratedPool()

			// Initialize lower, upper, and current ticks
			s.initializeTick(s.Ctx, test.currentTickIndex.Int64(), test.lowerTick.tickIndex, sdk.ZeroDec(), cl.EmptyCoins, test.lowerTick.uptimeTrackers, true)
			s.initializeTick(s.Ctx, test.currentTickIndex.Int64(), test.upperTick.tickIndex, sdk.ZeroDec(), cl.EmptyCoins, test.upperTick.uptimeTrackers, false)
			clPool.SetCurrentTick(test.currentTickIndex)
			s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, clPool)

			// Initialize global uptime accums
			addToUptimeAccums(s.Ctx, clPool.GetId(), s.App.ConcentratedLiquidityKeeper, test.globalUptimeAccumValues)

			// If applicable, set up existing position and update ticks & global accums
			if test.existingPosition {
				err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePositionUptime(s.Ctx, clPool.GetId(), test.positionLiquidity, s.TestAccs[0], test.lowerTick.tickIndex, test.upperTick.tickIndex, test.positionLiquidity, DefaultJoinTime, DefaultPositionId)
				s.Require().NoError(err)
				s.App.ConcentratedLiquidityKeeper.SetPosition(s.Ctx, clPool.GetId(), s.TestAccs[0], test.lowerTick.tickIndex, test.upperTick.tickIndex, DefaultJoinTime, test.positionLiquidity, DefaultPositionId)

				s.initializeTick(s.Ctx, test.currentTickIndex.Int64(), test.newLowerTick.tickIndex, sdk.ZeroDec(), cl.EmptyCoins, test.newLowerTick.uptimeTrackers, true)
				s.initializeTick(s.Ctx, test.currentTickIndex.Int64(), test.newUpperTick.tickIndex, sdk.ZeroDec(), cl.EmptyCoins, test.newUpperTick.uptimeTrackers, false)
				clPool.SetCurrentTick(test.currentTickIndex)
				s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, clPool)

				addToUptimeAccums(s.Ctx, clPool.GetId(), s.App.ConcentratedLiquidityKeeper, test.addToGlobalAccums)
			}

			// --- System under test ---

			err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePositionUptime(s.Ctx, clPool.GetId(), test.positionLiquidity, s.TestAccs[0], test.lowerTick.tickIndex, test.upperTick.tickIndex, test.positionLiquidity, DefaultJoinTime, DefaultPositionId)

			// --- Error catching ---

			if test.expectedErr != nil {
				s.Require().Error(err)
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
				recordExists, err := uptimeAccums[uptimeIndex].HasPosition(positionName)
				s.Require().NoError(err)

				s.Require().True(recordExists)

				// Ensure position's record has correct values
				positionRecord, err := accum.GetPosition(uptimeAccums[uptimeIndex], positionName)
				s.Require().NoError(err)

				s.Require().Equal(test.expectedInitAccumValue[uptimeIndex], positionRecord.InitAccumValue)

				if test.existingPosition {
					s.Require().Equal(sdk.NewDec(2).Mul(test.positionLiquidity), positionRecord.NumShares)
				} else {
					s.Require().Equal(test.positionLiquidity, positionRecord.NumShares)
				}

				// Note that the rewards only apply to the initial shares, not the new ones
				s.Require().Equal(test.expectedUnclaimedRewards[uptimeIndex].MulDec(test.positionLiquidity), positionRecord.UnclaimedRewards)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCollectIncentives() {
	ownerWithValidPosition := s.TestAccs[0]
	uptimeHelper := getExpectedUptimes()
	oneDay := time.Hour * 24
	oneWeek := 7 * time.Hour * 24
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

	tests := map[string]struct {
		// setup parameters
		existingAccumLiquidity   []sdk.Dec
		addedUptimeGrowthInside  []sdk.DecCoins
		addedUptimeGrowthOutside []sdk.DecCoins
		currentTick              int64
		isInvalidPoolIdGiven     bool
		numPositions             int

		// inputs parameters
		positionParams positionParameters
		timeInPosition time.Duration

		// expectations
		expectedIncentivesClaimed sdk.Coins
		expectedError             error
	}{
		// ---Cases for lowerTick < currentTick < upperTick---

		"(lower < curr < upper) no uptime growth inside or outside range, 1D time in position": {
			currentTick: 1,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            oneDay,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth outside range but not inside, 1D time in position": {
			currentTick:              1,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneDay,
			// Since there was no growth inside the range, we expect no incentives to be claimed
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth inside range but not outside, 1D time in position": {
			currentTick:             1,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
		},
		"(lower < curr < upper) uptime growth both inside and outside range, 1D time in position": {
			currentTick:              1,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for. At the same time, growth outside does not affect the current position's incentive rewards.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
		},
		"(lower < curr < upper) no uptime growth inside or outside range, 1W time in position": {
			currentTick: 1,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            oneWeek,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth outside range but not inside, 1W time in position": {
			currentTick:              1,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneWeek,
			// Since there was no growth inside the range, we expect no incentives to be claimed
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth inside range but not outside, 1W time in position": {
			currentTick:             1,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneWeek,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneWeek, defaultMultiplier),
		},
		"(lower < curr < upper) uptime growth both inside and outside range, 1W time in position": {
			currentTick:              1,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneWeek,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for. At the same time, growth outside does not affect the current position's incentive rewards.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneWeek, defaultMultiplier),
		},
		"(lower < curr < upper) no uptime growth inside or outside range, no time in position": {
			currentTick: 1,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            0,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth outside range but not inside, no time in position": {
			currentTick:              1,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            0,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth inside range but not outside, no time in position": {
			currentTick:             1,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            0,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < curr < upper) uptime growth both inside and outside range, no time in position": {
			currentTick:              1,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            0,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},

		// ---Cases for currentTick < lowerTick < upperTick---

		"(curr < lower < upper) no uptime growth inside or outside range, 1D time in position": {
			currentTick: 0,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            oneDay,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth outside range but not inside, 1D time in position": {
			currentTick:              0,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneDay,
			// Since there was no growth inside the range, we expect no incentives to be claimed
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth inside range but not outside, 1D time in position": {
			currentTick:             0,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
		},
		"(curr < lower < upper) uptime growth both inside and outside range, 1D time in position": {
			currentTick:              0,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for. At the same time, growth outside does not affect the current position's incentive rewards.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
		},
		"(curr < lower < upper) no uptime growth inside or outside range, 1W time in position": {
			currentTick: 0,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            oneWeek,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth outside range but not inside, 1W time in position": {
			currentTick:              0,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneWeek,
			// Since there was no growth inside the range, we expect no incentives to be claimed
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth inside range but not outside, 1W time in position": {
			currentTick:             0,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneWeek,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneWeek, defaultMultiplier),
		},
		"(curr < lower < upper) uptime growth both inside and outside range, 1W time in position": {
			currentTick:              0,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneWeek,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for. At the same time, growth outside does not affect the current position's incentive rewards.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneWeek, defaultMultiplier),
		},
		"(curr < lower < upper) no uptime growth inside or outside range, no time in position": {
			currentTick: 0,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            0,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth outside range but not inside, no time in position": {
			currentTick:              0,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            0,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth inside range but not outside, no time in position": {
			currentTick:             0,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            0,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(curr < lower < upper) uptime growth both inside and outside range, no time in position": {
			currentTick:              0,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            0,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},

		// ---Cases for lowerTick < upperTick < currentTick---

		"(lower < upper < curr) no uptime growth inside or outside range, 1D time in position": {
			currentTick: 3,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            oneDay,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth outside range but not inside, 1D time in position": {
			currentTick:              3,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneDay,
			// Since there was no growth inside the range, we expect no incentives to be claimed
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth inside range but not outside, 1D time in position": {
			currentTick:             3,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
		},
		"(lower < upper < curr) uptime growth both inside and outside range, 1D time in position": {
			currentTick:              3,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for. At the same time, growth outside does not affect the current position's incentive rewards.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
		},
		"(lower < upper < curr) no uptime growth inside or outside range, 1W time in position": {
			currentTick: 3,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            oneWeek,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth outside range but not inside, 1W time in position": {
			currentTick:              3,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneWeek,
			// Since there was no growth inside the range, we expect no incentives to be claimed
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth inside range but not outside, 1W time in position": {
			currentTick:             3,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneWeek,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneWeek, defaultMultiplier),
		},
		"(lower < upper < curr) uptime growth both inside and outside range, 1W time in position": {
			currentTick:              3,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneWeek,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneWeek, defaultMultiplier),
		},
		"(lower < upper < curr) no uptime growth inside or outside range, no time in position": {
			currentTick: 3,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            0,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth outside range but not inside, no time in position": {
			currentTick:              3,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            0,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth inside range but not outside, no time in position": {
			currentTick:             3,
			addedUptimeGrowthInside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            0,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},
		"(lower < upper < curr) uptime growth both inside and outside range, no time in position": {
			currentTick:              3,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:              1,
			timeInPosition:            0,
			expectedIncentivesClaimed: sdk.Coins(nil),
		},

		// Edge case tests

		"(curr = lower) uptime growth both inside and outside range, 1D time in position": {
			currentTick:              0,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneDay,
			// We expect this case to behave like (lower < curr < upper)
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
		},
		"(curr = upper) uptime growth both inside and outside range, 1D time in position": {
			currentTick:              2,
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   1,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneDay,
			// We expect this case to behave like (lower < upper < curr)
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
		},
		"other liquidity on uptime accums: (lower < curr < upper) uptime growth both inside and outside range, 1D time in position": {
			currentTick: 1,
			existingAccumLiquidity: []sdk.Dec{
				sdk.NewDec(99900123432),
				sdk.NewDec(18942),
				sdk.NewDec(0),
				sdk.NewDec(9981),
				sdk.NewDec(1),
			},
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   1,
			timeInPosition: oneDay,
			// Since there is no other existing liquidity, we expect all of the growth inside to accrue to be claimed for the
			// uptimes the position qualifies for.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
		},
		"multiple positions in same range: (lower < curr < upper) uptime growth both inside and outside range, 1D time in position": {
			currentTick: 1,
			existingAccumLiquidity: []sdk.Dec{
				sdk.NewDec(99900123432),
				sdk.NewDec(18942),
				sdk.NewDec(0),
				sdk.NewDec(9981),
				sdk.NewDec(1),
			},
			addedUptimeGrowthInside:  uptimeHelper.hundredTokensMultiDenom,
			addedUptimeGrowthOutside: uptimeHelper.hundredTokensMultiDenom,
			positionParams: positionParameters{
				owner:       ownerWithValidPosition,
				lowerTick:   0,
				upperTick:   2,
				liquidity:   DefaultLiquidityAmt,
				joinTime:    defaultJoinTime,
				positionId:  DefaultPositionId,
				collectTime: defaultJoinTime.Add(100),
			},
			numPositions:   3,
			timeInPosition: oneDay,
			// Since we introduced positionIDs, despite these position having the same range and pool, only
			// the position ID being claimed will be considered for the claim.
			expectedIncentivesClaimed: expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, oneDay, defaultMultiplier),
		},

		// Error catching

		"position does not exist": {
			currentTick: 1,

			numPositions: 0,

			expectedIncentivesClaimed: sdk.Coins{},
			expectedError:             cltypes.PositionIdNotFoundError{PositionId: DefaultPositionId},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()

			// We fix join time so tests are deterministic
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)

			validPool := s.PrepareConcentratedPool()
			validPoolId := validPool.GetId()

			s.FundAcc(validPool.GetIncentivesAddress(), tc.expectedIncentivesClaimed)

			clKeeper := s.App.ConcentratedLiquidityKeeper
			ctx := s.Ctx

			if tc.numPositions > 0 {
				// Initialize lower and upper ticks with empty uptime trackers
				s.initializeTick(ctx, tc.currentTick, tc.positionParams.lowerTick, tc.positionParams.liquidity, cl.EmptyCoins, wrapUptimeTrackers(uptimeHelper.emptyExpectedAccumValues), true)
				s.initializeTick(ctx, tc.currentTick, tc.positionParams.upperTick, tc.positionParams.liquidity, cl.EmptyCoins, wrapUptimeTrackers(uptimeHelper.emptyExpectedAccumValues), false)

				if tc.existingAccumLiquidity != nil {
					s.addLiquidityToUptimeAccumulators(ctx, validPoolId, tc.existingAccumLiquidity, tc.positionParams.positionId+1)
				}

				// Initialize all positions
				for i := 0; i < tc.numPositions; i++ {
					err := clKeeper.InitOrUpdatePosition(ctx, validPoolId, ownerWithValidPosition, tc.positionParams.lowerTick, tc.positionParams.upperTick, tc.positionParams.liquidity, tc.positionParams.joinTime, uint64(i+1))
					s.Require().NoError(err)
				}
				ctx = ctx.WithBlockTime(ctx.BlockTime().Add(tc.timeInPosition))

				// Add to uptime growth inside range
				if tc.addedUptimeGrowthInside != nil {
					s.addUptimeGrowthInsideRange(ctx, validPoolId, ownerWithValidPosition, tc.currentTick, tc.positionParams.lowerTick, tc.positionParams.upperTick, tc.addedUptimeGrowthInside)
				}

				// Add to uptime growth outside range
				if tc.addedUptimeGrowthOutside != nil {
					s.addUptimeGrowthOutsideRange(ctx, validPoolId, ownerWithValidPosition, tc.currentTick, tc.positionParams.lowerTick, tc.positionParams.upperTick, tc.addedUptimeGrowthOutside)
				}
			}

			validPool.SetCurrentTick(sdk.NewInt(tc.currentTick))
			clKeeper.SetPool(ctx, validPool)

			// Checkpoint starting balance to compare against later
			poolBalanceBeforeCollect := s.App.BankKeeper.GetAllBalances(ctx, validPool.GetAddress())
			incentivesBalanceBeforeCollect := s.App.BankKeeper.GetAllBalances(ctx, validPool.GetIncentivesAddress())
			ownerBalancerBeforeCollect := s.App.BankKeeper.GetAllBalances(ctx, ownerWithValidPosition)

			// Set up invalid pool ID for error-catching case(s)
			sutPoolId := validPoolId
			if tc.isInvalidPoolIdGiven {
				sutPoolId = sutPoolId + 1
			}

			// System under test
			actualIncentivesClaimed, err := clKeeper.CollectIncentives(ctx, ownerWithValidPosition, DefaultPositionId)

			// Assertions

			poolBalanceAfterCollect := s.App.BankKeeper.GetAllBalances(ctx, validPool.GetAddress())
			incentivesBalanceAfterCollect := s.App.BankKeeper.GetAllBalances(ctx, validPool.GetIncentivesAddress())
			ownerBalancerAfterCollect := s.App.BankKeeper.GetAllBalances(ctx, ownerWithValidPosition)

			// Ensure pool balances are unchanged independent of error.
			s.Require().Equal(poolBalanceBeforeCollect, poolBalanceAfterCollect)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				s.Require().Equal(tc.expectedIncentivesClaimed, actualIncentivesClaimed)

				// Ensure balances are unchanged
				s.Require().Equal(incentivesBalanceBeforeCollect, incentivesBalanceAfterCollect)
				s.Require().Equal(ownerBalancerAfterCollect, ownerBalancerBeforeCollect)
				return
			}

			// Ensure claimed amount is correct
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedIncentivesClaimed.String(), actualIncentivesClaimed.String())

			// Ensure balances are updated by the correct amounts
			s.Require().Equal(tc.expectedIncentivesClaimed.String(), (incentivesBalanceBeforeCollect.Sub(incentivesBalanceAfterCollect)).String())
			s.Require().Equal(tc.expectedIncentivesClaimed.String(), (ownerBalancerAfterCollect.Sub(ownerBalancerBeforeCollect)).String())
		})
	}
}

func (s *KeeperTestSuite) TestCreateIncentive() {
	type testCreateIncentive struct {
		poolId             uint64
		isInvalidPoolId    bool
		sender             sdk.AccAddress
		senderBalance      sdk.Coins
		recordToSet        types.IncentiveRecord
		existingRecords    []types.IncentiveRecord
		minimumGasConsumed uint64

		expectedError error
	}
	tests := map[string]testCreateIncentive{
		"valid incentive record": {
			poolId: defaultPoolId,
			sender: sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveDenom,
					incentiveRecordOne.IncentiveRecordBody.RemainingAmount.Ceil().RoundInt(),
				),
			),
			recordToSet:        incentiveRecordOne,
			minimumGasConsumed: uint64(0),
		},
		"record with different denom, emission rate, and min uptime": {
			poolId: defaultPoolId,
			sender: sdk.MustAccAddressFromBech32(incentiveRecordTwo.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordTwo.IncentiveDenom,
					incentiveRecordTwo.IncentiveRecordBody.RemainingAmount.Ceil().RoundInt(),
				),
			),
			recordToSet:        incentiveRecordTwo,
			minimumGasConsumed: uint64(0),
		},
		"record with different start time": {
			poolId: defaultPoolId,
			sender: sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveDenom,
					incentiveRecordOne.IncentiveRecordBody.RemainingAmount.Ceil().RoundInt(),
				),
			),
			recordToSet:        withStartTime(incentiveRecordOne, defaultStartTime.Add(time.Hour)),
			minimumGasConsumed: uint64(0),
		},
		"record with different incentive amount": {
			poolId: defaultPoolId,
			sender: sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveDenom,
					sdk.NewInt(8),
				),
			),
			recordToSet:        withAmount(incentiveRecordOne, sdk.NewDec(8)),
			minimumGasConsumed: uint64(0),
		},
		"existing incentive records on different uptime accumulators": {
			poolId: defaultPoolId,
			sender: sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveDenom,
					incentiveRecordOne.IncentiveRecordBody.RemainingAmount.Ceil().RoundInt(),
				),
			),
			recordToSet:     incentiveRecordOne,
			existingRecords: []types.IncentiveRecord{incentiveRecordTwo, incentiveRecordThree},

			// We still expect a minimum of 0 since the existing records are on other uptime accumulators
			minimumGasConsumed: uint64(0),
		},
		"existing incentive records on the same uptime accumulator": {
			poolId: defaultPoolId,
			sender: sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveDenom,
					incentiveRecordOne.IncentiveRecordBody.RemainingAmount.Ceil().RoundInt(),
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

		// Error catching
		"pool doesn't exist": {
			isInvalidPoolId: true,

			poolId: defaultPoolId,
			sender: sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveDenom,
					incentiveRecordOne.IncentiveRecordBody.RemainingAmount.Ceil().RoundInt(),
				),
			),
			recordToSet: incentiveRecordOne,

			expectedError: types.PoolNotFoundError{PoolId: 2},
		},
		"zero incentive amount": {
			poolId: defaultPoolId,
			sender: sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveDenom,
					sdk.ZeroInt(),
				),
			),
			recordToSet: withAmount(incentiveRecordOne, sdk.ZeroDec()),

			expectedError: types.NonPositiveIncentiveAmountError{PoolId: 1, IncentiveAmount: sdk.ZeroDec()},
		},
		"negative incentive amount": {
			poolId: defaultPoolId,
			sender: sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveDenom,
					sdk.ZeroInt(),
				),
			),
			recordToSet: withAmount(incentiveRecordOne, sdk.NewDec(-1)),

			expectedError: types.NonPositiveIncentiveAmountError{PoolId: 1, IncentiveAmount: sdk.NewDec(-1)},
		},
		"start time too early": {
			poolId: defaultPoolId,
			sender: sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveDenom,
					incentiveRecordOne.IncentiveRecordBody.RemainingAmount.Ceil().RoundInt(),
				),
			),
			recordToSet: withStartTime(incentiveRecordOne, defaultBlockTime.Add(-1*time.Second)),

			expectedError: types.StartTimeTooEarlyError{PoolId: 1, CurrentBlockTime: defaultBlockTime, StartTime: defaultBlockTime.Add(-1 * time.Second)},
		},
		"zero emission rate": {
			poolId: defaultPoolId,
			sender: sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveDenom,
					incentiveRecordOne.IncentiveRecordBody.RemainingAmount.Ceil().RoundInt(),
				),
			),
			recordToSet: withEmissionRate(incentiveRecordOne, sdk.ZeroDec()),

			expectedError: types.NonPositiveEmissionRateError{PoolId: 1, EmissionRate: sdk.ZeroDec()},
		},
		"negative emission rate": {
			poolId: defaultPoolId,
			sender: sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveDenom,
					incentiveRecordOne.IncentiveRecordBody.RemainingAmount.Ceil().RoundInt(),
				),
			),
			recordToSet: withEmissionRate(incentiveRecordOne, sdk.NewDec(-1)),

			expectedError: types.NonPositiveEmissionRateError{PoolId: 1, EmissionRate: sdk.NewDec(-1)},
		},
		"unsupported min uptime": {
			poolId: defaultPoolId,
			sender: sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(
				sdk.NewCoin(
					incentiveRecordOne.IncentiveDenom,
					incentiveRecordOne.IncentiveRecordBody.RemainingAmount.Ceil().RoundInt(),
				),
			),
			recordToSet: withMinUptime(incentiveRecordOne, time.Hour*3),

			expectedError: types.InvalidMinUptimeError{PoolId: 1, MinUptime: time.Hour * 3, SupportedUptimes: types.SupportedUptimes},
		},
		"insufficient sender balance": {
			poolId:        defaultPoolId,
			sender:        sdk.MustAccAddressFromBech32(incentiveRecordOne.IncentiveCreatorAddr),
			senderBalance: sdk.NewCoins(),
			recordToSet:   incentiveRecordOne,

			expectedError: types.IncentiveInsufficientBalanceError{PoolId: 1, IncentiveDenom: incentiveRecordOne.IncentiveDenom, IncentiveAmount: incentiveRecordOne.IncentiveRecordBody.RemainingAmount.Ceil().RoundInt()},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()

			// We fix blocktime to ensure tests are deterministic
			s.Ctx = s.Ctx.WithBlockTime(defaultBlockTime)

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

			// system under test
			incentiveRecord, err := clKeeper.CreateIncentive(s.Ctx, tc.poolId, tc.sender, tc.recordToSet.IncentiveDenom, tc.recordToSet.IncentiveRecordBody.RemainingAmount.Ceil().RoundInt(), tc.recordToSet.IncentiveRecordBody.EmissionRate, tc.recordToSet.IncentiveRecordBody.StartTime, tc.recordToSet.MinUptime)

			// Assertions
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())

				// Ensure nothing was placed in state
				recordInState, err := clKeeper.GetIncentiveRecord(s.Ctx, tc.poolId, tc.recordToSet.IncentiveDenom, tc.recordToSet.MinUptime, tc.sender)
				s.Require().Error(err)
				s.Require().Equal(types.IncentiveRecord{}, recordInState)

				return
			}
			s.Require().NoError(err)

			// Returned incentive record should equal both to what's in state and what we expect
			recordInState, err := clKeeper.GetIncentiveRecord(s.Ctx, tc.poolId, tc.recordToSet.IncentiveDenom, tc.recordToSet.MinUptime, tc.sender)
			s.Require().Equal(tc.recordToSet, recordInState)
			s.Require().Equal(tc.recordToSet, incentiveRecord)

			// Ensure that at least the minimum amount of gas was charged (based on number of existing incentives for current uptime)
			gasConsumed := s.Ctx.GasMeter().GasConsumed() - existingGasConsumed
			s.Require().True(gasConsumed >= tc.minimumGasConsumed)

			// Ensure that existing records aren't affected
			for _, incentiveRecord := range tc.existingRecords {
				_, err := clKeeper.GetIncentiveRecord(s.Ctx, tc.poolId, incentiveRecord.IncentiveDenom, incentiveRecord.MinUptime, sdk.MustAccAddressFromBech32(incentiveRecord.IncentiveCreatorAddr))
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestPrepareAccumAndClaimRewards() {
	validPositionKey := cltypes.KeyFeePositionAccumulator(1)
	invalidPositionKey := cltypes.KeyFeePositionAccumulator(2)
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
	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			// Setup test env.
			s.SetupTest()
			s.PrepareConcentratedPool()
			clKeeper := s.App.ConcentratedLiquidityKeeper

			poolFeeAccumulator, err := clKeeper.GetFeeAccumulator(s.Ctx, defaultPoolId)
			s.Require().NoError(err)
			positionKey := validPositionKey

			// Initialize position accumulator.
			err = poolFeeAccumulator.NewPositionCustomAcc(positionKey, sdk.OneDec(), sdk.DecCoins{}, nil)
			s.Require().NoError(err)

			// Record the initial position accumulator value.
			positionPre, err := accum.GetPosition(poolFeeAccumulator, positionKey)
			s.Require().NoError(err)

			// If the test case requires an invalid position key, set it.
			if tc.invalidPositionKey {
				positionKey = invalidPositionKey
			}

			poolFeeAccumulator.AddToAccumulator(tc.growthOutside.Add(tc.growthInside...))

			// System under test.
			amountClaimed, _, err := cl.PrepareAccumAndClaimRewards(poolFeeAccumulator, positionKey, tc.growthOutside)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)

			// We expect claimed rewards to be equal to growth inside
			expectedCoins := sdk.NormalizeCoins(tc.growthInside)
			s.Require().Equal(expectedCoins, amountClaimed)

			// Record the final position accumulator value.
			positionPost, err := accum.GetPosition(poolFeeAccumulator, positionKey)
			s.Require().NoError(err)

			// Check that the difference between the new and old position accumulator values is equal to the growth inside (since
			// we recalibrate the position accum value after claiming).
			positionAccumDelta := positionPost.InitAccumValue.Sub(positionPre.InitAccumValue)
			s.Require().Equal(tc.growthInside, positionAccumDelta)
		})
	}
}

// Note that the non-forfeit cases are thoroughly tested in `TestCollectIncentives`
func (s *KeeperTestSuite) TestClaimAllIncentives() {
	uptimeHelper := getExpectedUptimes()
	defaultSender := s.TestAccs[0]
	tests := map[string]struct {
		name              string
		poolId            uint64
		positionIdCreate  uint64
		positionIdClaim   uint64
		defaultJoinTime   bool
		growthInside      []sdk.DecCoins
		growthOutside     []sdk.DecCoins
		forfeitIncentives bool
		expectedError     error
	}{
		"happy path: claim rewards without forfeiting": {
			poolId:           validPoolId,
			positionIdCreate: DefaultPositionId,
			positionIdClaim:  DefaultPositionId,
			defaultJoinTime:  true,
			growthInside:     uptimeHelper.hundredTokensMultiDenom,
			growthOutside:    uptimeHelper.twoHundredTokensMultiDenom,
		},
		"claim and forfeit rewards": {
			poolId:            validPoolId,
			positionIdCreate:  DefaultPositionId,
			positionIdClaim:   DefaultPositionId,
			defaultJoinTime:   true,
			growthInside:      uptimeHelper.hundredTokensMultiDenom,
			growthOutside:     uptimeHelper.twoHundredTokensMultiDenom,
			forfeitIncentives: true,
		},
		"claim and forfeit rewards when no rewards have accrued": {
			poolId:            validPoolId,
			positionIdCreate:  DefaultPositionId,
			positionIdClaim:   DefaultPositionId,
			defaultJoinTime:   true,
			forfeitIncentives: true,
		},
		"claim and forfeit rewards with varying amounts and different denoms": {
			poolId:            validPoolId,
			positionIdCreate:  DefaultPositionId,
			positionIdClaim:   DefaultPositionId,
			defaultJoinTime:   true,
			growthInside:      uptimeHelper.varyingTokensMultiDenom,
			growthOutside:     uptimeHelper.varyingTokensSingleDenom,
			forfeitIncentives: true,
		},

		// error catching

		"error: non existent position": {
			poolId:           validPoolId + 1,
			positionIdCreate: DefaultPositionId,
			positionIdClaim:  DefaultPositionId + 1, // non existent position
			defaultJoinTime:  true,
			growthInside:     uptimeHelper.hundredTokensMultiDenom,
			growthOutside:    uptimeHelper.twoHundredTokensMultiDenom,

			expectedError: cltypes.PositionIdNotFoundError{PositionId: DefaultPositionId + 1},
		},

		"error: negative duration": {
			poolId:           validPoolId,
			positionIdCreate: DefaultPositionId,
			positionIdClaim:  DefaultPositionId,
			defaultJoinTime:  false,
			growthInside:     uptimeHelper.hundredTokensMultiDenom,
			growthOutside:    uptimeHelper.twoHundredTokensMultiDenom,

			expectedError: cltypes.NegativeDurationError{Duration: time.Hour * 504 * -1},
		},
	}
	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			// --- Setup test env ---

			s.SetupTest()
			clPool := s.PrepareConcentratedPool()
			clKeeper := s.App.ConcentratedLiquidityKeeper

			joinTime := s.Ctx.BlockTime()
			if !tc.defaultJoinTime {
				joinTime = joinTime.AddDate(0, 0, 28)
			}

			// Initialize position
			err := clKeeper.InitOrUpdatePosition(s.Ctx, validPoolId, defaultSender, DefaultLowerTick, DefaultUpperTick, sdk.OneDec(), joinTime, tc.positionIdCreate)
			s.Require().NoError(err)

			clPool.SetCurrentTick(DefaultCurrTick)
			if tc.growthOutside != nil {
				s.addUptimeGrowthOutsideRange(s.Ctx, validPoolId, defaultSender, DefaultCurrTick.Int64(), DefaultLowerTick, DefaultUpperTick, tc.growthOutside)
			}

			if tc.growthInside != nil {
				s.addUptimeGrowthInsideRange(s.Ctx, validPoolId, defaultSender, DefaultCurrTick.Int64(), DefaultLowerTick, DefaultUpperTick, tc.growthInside)
			}

			err = clKeeper.SetPool(s.Ctx, clPool)
			s.Require().NoError(err)

			// Store initial accum values for comparison later
			initUptimeAccumValues, err := clKeeper.GetUptimeAccumulatorValues(s.Ctx, validPoolId)
			s.Require().NoError(err)

			// Store initial pool and sender balances for comparison later
			initSenderBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, defaultSender)
			initPoolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, clPool.GetAddress())

			if !tc.forfeitIncentives {
				// Let enough time elapse for the position to accrue rewards for all uptimes
				s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(types.SupportedUptimes[len(types.SupportedUptimes)-1]))
			}

			// --- System under test ---

			amountClaimed, amountForfeited, err := clKeeper.ClaimAllIncentivesForPosition(s.Ctx, tc.positionIdClaim)

			// --- Assertions ---

			// Pull new balances for comparison
			newSenderBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, defaultSender)
			newPoolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, clPool.GetAddress())

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectedError)

				// Ensure balances have not been mutated
				s.Require().Equal(initSenderBalances, newSenderBalances)
				s.Require().Equal(initPoolBalances, newPoolBalances)
				return
			}
			s.Require().NoError(err)

			// Ensure that forfeited incentives were properly added to their respective accumulators
			if tc.forfeitIncentives {
				newUptimeAccumValues, err := clKeeper.GetUptimeAccumulatorValues(s.Ctx, validPoolId)
				s.Require().NoError(err)

				// Subtract the initial accum values to get the delta
				uptimeAccumDeltaValues, err := osmoutils.SubDecCoinArrays(newUptimeAccumValues, initUptimeAccumValues)
				s.Require().NoError(err)

				// Convert DecCoins to Coins by truncation for comparison
				normalizedUptimeAccumDelta := sdk.NewCoins()
				for _, uptimeAccumDelta := range uptimeAccumDeltaValues {
					normalizedUptimeAccumDelta = normalizedUptimeAccumDelta.Add(sdk.NormalizeCoins(uptimeAccumDelta)...)
				}

				s.Require().Equal(normalizedUptimeAccumDelta.String(), amountClaimed.Add(amountForfeited...).String())
			} else {
				// We expect claimed rewards to be equal to growth inside
				expectedCoins := sdk.Coins(nil)
				for _, growthInside := range tc.growthInside {
					expectedCoins = expectedCoins.Add(sdk.NormalizeCoins(growthInside)...)
				}
				s.Require().Equal(expectedCoins, amountClaimed.Add(amountForfeited...))
			}

			// Ensure balances have not been mutated
			s.Require().Equal(initSenderBalances, newSenderBalances)
			s.Require().Equal(initPoolBalances, newPoolBalances)
		})
	}
}

func (s *KeeperTestSuite) TestGetAllIncentivesForUptime() {
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

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())

				s.Require().Equal(tc.expectedRecords, retrievedRecords)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedRecords, retrievedRecords)

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
}

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
	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			retrievedUptimeIndex, err := cl.FindUptimeIndex(tc.requestedUptime)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				s.Require().Equal(tc.expectedUptimeIndex, retrievedUptimeIndex)

				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedUptimeIndex, retrievedUptimeIndex)
		})
	}
}
