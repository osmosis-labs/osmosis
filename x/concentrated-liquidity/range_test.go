package concentrated_liquidity_test

import (
	"fmt"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

type RangeTestParams struct {
	// -- Base amounts --

	// Base number of assets for each position
	baseAssets sdk.Coins
	// Base number of positions for each range
	baseNumPositions int
	// Base amount to swap for each swap
	baseSwapAmount osmomath.Int
	// Base amount to add after each new position
	baseTimeBetweenJoins time.Duration
	// Base incentive amount to have on each incentive record
	baseIncentiveAmount osmomath.Int
	// Base emission rate per second for incentive
	baseEmissionRate osmomath.Dec
	// Base denom for each incentive record (ID appended to this)
	baseIncentiveDenom string
	// List of addresses to swap from (randomly selected for each swap)
	numSwapAddresses int

	// -- Pool params --

	spreadFactor osmomath.Dec
	tickSpacing  uint64

	// -- Fuzz params --

	fuzzAssets           bool
	fuzzNumPositions     bool
	fuzzSwapAmounts      bool
	fuzzTimeBetweenJoins bool
	fuzzIncentiveRecords bool

	// -- Optional additional test dimensions --

	// Have a single address for all positions in each range
	singleAddrPerRange bool
	// Create new active incentive records between each join
	newActiveIncentivesBetweenJoins bool
	// Create new inactive incentive records between each join
	newInactiveIncentivesBetweenJoins bool
	// Fund each position address with double the expected amount of assets.
	// Should only be used for cases where join amount gets pushed up due to
	// precision near min tick.
	doubleFundPositionAddr bool
	// Adjust input amounts for first position to set the starting current tick
	// to the given value.
	startingCurrentTick int64
}

var (
	DefaultRangeTestParams = RangeTestParams{
		// Base amounts
		baseNumPositions:     10,
		baseAssets:           sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(5000000000)), sdk.NewCoin(USDC, osmomath.NewInt(5000000000))),
		baseTimeBetweenJoins: time.Hour,
		baseSwapAmount:       osmomath.NewInt(10000000),
		numSwapAddresses:     10,
		baseIncentiveAmount:  osmomath.NewInt(1000000000000000000),
		baseEmissionRate:     osmomath.NewDec(1),
		baseIncentiveDenom:   "incentiveDenom",

		// Pool params
		spreadFactor: DefaultSpreadFactor,
		tickSpacing:  uint64(1),

		// Fuzz params
		fuzzNumPositions:     true,
		fuzzAssets:           true,
		fuzzSwapAmounts:      true,
		fuzzTimeBetweenJoins: true,
	}
	RangeTestParamsLargeSwap = RangeTestParams{
		// Base amounts
		baseNumPositions:     10,
		baseAssets:           sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(5000000000)), sdk.NewCoin(USDC, osmomath.NewInt(5000000000))),
		baseTimeBetweenJoins: time.Hour,
		baseSwapAmount:       osmomath.Int(osmomath.MustNewDecFromStr("100000000000000000000000000000000000000")),
		numSwapAddresses:     10,
		baseIncentiveAmount:  osmomath.NewInt(1000000000000000000),
		baseEmissionRate:     osmomath.NewDec(1),
		baseIncentiveDenom:   "incentiveDenom",

		// Pool params
		spreadFactor: DefaultSpreadFactor,
		tickSpacing:  uint64(100),

		// Fuzz params
		fuzzNumPositions:     true,
		fuzzAssets:           true,
		fuzzTimeBetweenJoins: true,
	}
	RangeTestParamsNoFuzzNoSwap = RangeTestParams{
		// Base amounts
		baseNumPositions:     1,
		baseAssets:           sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(5000000000)), sdk.NewCoin(USDC, osmomath.NewInt(5000000000))),
		baseTimeBetweenJoins: time.Hour,
		baseIncentiveAmount:  osmomath.NewInt(1000000000000000000),
		baseEmissionRate:     osmomath.NewDec(1),
		baseIncentiveDenom:   "incentiveDenom",

		// Pool params
		spreadFactor: DefaultSpreadFactor,
		tickSpacing:  uint64(1),
	}
)

func withDoubleFundedLP(params RangeTestParams) RangeTestParams {
	params.doubleFundPositionAddr = true
	return params
}

func withCurrentTick(params RangeTestParams, tick int64) RangeTestParams {
	params.startingCurrentTick = tick
	return params
}

func withTickSpacing(params RangeTestParams, tickSpacing uint64) RangeTestParams {
	params.tickSpacing = tickSpacing
	return params
}

func withNoSwap(params RangeTestParams) RangeTestParams {
	params.baseSwapAmount = osmomath.Int{}
	return params
}

// setupRangesAndAssertInvariants sets up the state specified by `testParams` on the given set of ranges.
// It also asserts global invariants at each intermediate step.
func (s *KeeperTestSuite) setupRangesAndAssertInvariants(pool types.ConcentratedPoolExtension, ranges [][]int64, testParams RangeTestParams) {

	// --- Parse test params ---

	// Prepare a slice tracking how many positions to create on each range.
	numPositionSlice, totalPositions := s.prepareNumPositionSlice(ranges, testParams.baseNumPositions, testParams.fuzzNumPositions)

	// Set up position accounts
	var positionAddresses []sdk.AccAddress
	if testParams.singleAddrPerRange {
		positionAddresses = apptesting.CreateRandomAccounts(len(ranges))
	} else {
		positionAddresses = apptesting.CreateRandomAccounts(totalPositions)
	}

	// Set up swap accounts
	swapAddresses := apptesting.CreateRandomAccounts(testParams.numSwapAddresses)

	// --- Incentive setup ---

	if testParams.baseIncentiveAmount != (osmomath.Int{}) {
		incentiveAddr := apptesting.CreateRandomAccounts(1)[0]
		incentiveAmt := testParams.baseIncentiveAmount
		emissionRate := testParams.baseEmissionRate
		incentiveCoin := sdk.NewCoin(fmt.Sprintf("%s%d", testParams.baseIncentiveDenom, 0), incentiveAmt)
		s.FundAcc(incentiveAddr, sdk.NewCoins(incentiveCoin))
		_, err := s.Clk.CreateIncentive(s.Ctx, pool.GetId(), incentiveAddr, incentiveCoin, emissionRate, s.Ctx.BlockTime(), types.DefaultAuthorizedUptimes[0])
		s.Require().NoError(err)
	}

	// --- Position setup ---

	// This loop runs through each given tick range and does the following at each iteration:
	// 1. Set up a position
	// 2. Let time elapse
	// 3. Execute a swap
	totalLiquidity, totalAssets, totalTimeElapsed, allPositionIds, lastVisitedBlockIndex, cumulativeEmittedIncentives, lastIncentiveTrackerUpdate := osmomath.ZeroDec(), sdk.NewCoins(), time.Duration(0), []uint64{}, 0, sdk.DecCoins{}, s.Ctx.BlockTime()
	for curRange := range ranges {
		curBlock := 0
		startNumPositions := len(allPositionIds)
		for curNumPositions := lastVisitedBlockIndex; curNumPositions < lastVisitedBlockIndex+numPositionSlice[curRange]; curNumPositions++ {
			// By default we create a new address for each position, but if the test params specify using a single address
			// for each range, we handle that logic here.
			var curAddr sdk.AccAddress
			if testParams.singleAddrPerRange {
				// If we are using a single address per range, we use the address corresponding to the current range.
				curAddr = positionAddresses[curRange]
			} else {
				// If we're not using a single address per range, we use a unique address for each position.
				curAddr = positionAddresses[curNumPositions]
			}

			// Set up assets for new position
			curAssets := getRandomizedAssets(testParams.baseAssets, testParams.fuzzAssets)

			// If a desired current tick was specified, retrieve special asset amounts for the first position
			if testParams.startingCurrentTick != 0 && curNumPositions == 0 {
				curAssets = s.getInitialPositionAssets(pool, testParams.startingCurrentTick)
			}

			roundingError := sdk.NewCoins(sdk.NewCoin(pool.GetToken0(), osmomath.OneInt()), sdk.NewCoin(pool.GetToken1(), osmomath.OneInt()))
			s.FundAcc(curAddr, curAssets.Add(roundingError...))

			// Double fund LP address if applicable
			if testParams.doubleFundPositionAddr {
				s.FundAcc(curAddr, curAssets.Add(roundingError...))
			}

			// TODO: implement intermediate record creation with fuzzing

			// Track emitted incentives here
			cumulativeEmittedIncentives, lastIncentiveTrackerUpdate = s.trackEmittedIncentives(cumulativeEmittedIncentives, lastIncentiveTrackerUpdate)

			// Set up position
			positionData, err := s.Clk.CreatePosition(s.Ctx, pool.GetId(), curAddr, curAssets, osmomath.ZeroInt(), osmomath.ZeroInt(), ranges[curRange][0], ranges[curRange][1])
			s.Require().NoError(err)

			// Ensure position was set up correctly and didn't break global invariants
			s.Require().Equal(ranges[curRange][0], positionData.LowerTick)
			s.Require().Equal(ranges[curRange][1], positionData.UpperTick)
			s.assertGlobalInvariants(ExpectedGlobalRewardValues{})

			// Let time elapse after join if applicable
			timeElapsed := s.addRandomizedBlockTime(testParams.baseTimeBetweenJoins, testParams.fuzzTimeBetweenJoins)

			// Execute swap against pool if applicable
			swappedIn, swappedOut := s.executeRandomizedSwap(pool, swapAddresses, testParams.baseSwapAmount, testParams.fuzzSwapAmounts)
			s.assertGlobalInvariants(ExpectedGlobalRewardValues{})

			// Track changes to state
			actualAddedCoins := sdk.NewCoins(sdk.NewCoin(pool.GetToken0(), positionData.Amount0), sdk.NewCoin(pool.GetToken1(), positionData.Amount1))
			totalAssets = totalAssets.Add(actualAddedCoins...)
			if testParams.baseSwapAmount != (osmomath.Int{}) {
				totalAssets = totalAssets.Add(swappedIn).Sub(sdk.NewCoins(swappedOut)...)
			}
			totalLiquidity = totalLiquidity.Add(positionData.Liquidity)
			totalTimeElapsed = totalTimeElapsed + timeElapsed
			allPositionIds = append(allPositionIds, positionData.ID)
			curBlock++
		}
		endNumPositions := len(allPositionIds)

		// Ensure the correct number of positions were set up in current range
		s.Require().Equal(numPositionSlice[curRange], endNumPositions-startNumPositions, "Incorrect number of positions set up in range %d", curRange)

		lastVisitedBlockIndex += curBlock
	}

	// Ensure that the correct number of positions were set up globally
	s.Require().Equal(totalPositions, len(allPositionIds))

	// Ensure the pool balance is exactly equal to the assets added + amount swapped in - amount swapped out
	poolAssets := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
	poolSpreadRewards := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetSpreadRewardsAddress())
	// We rebuild coins to handle nil cases cleanly
	s.Require().Equal(sdk.NewCoins(totalAssets...), sdk.NewCoins(poolAssets.Add(poolSpreadRewards...)...))

	// Do a final checkpoint for incentives and then run assertions on expected global claimable value
	cumulativeEmittedIncentives, lastIncentiveTrackerUpdate = s.trackEmittedIncentives(cumulativeEmittedIncentives, lastIncentiveTrackerUpdate)
	truncatedEmissions, _ := cumulativeEmittedIncentives.TruncateDecimal()

	// Run global assertions with an optional parameter specifying the expected incentive amount claimable by all positions.
	// We specifically need to do this for incentives because all the emissions are pre-loaded into the incentive address, making
	// balance assertions pass trivially in most cases.
	s.assertGlobalInvariants(ExpectedGlobalRewardValues{TotalIncentives: truncatedEmissions})
}

// numPositionSlice prepares a slice tracking the number of positions to create on each range, fuzzing the number at each step if applicable.
// Returns a slice representing the number of positions for each range index.
//
// We run this logic in a separate function for two main reasons:
// 1. Simplify position setup logic by fuzzing the number of positions upfront, letting us loop through the positions to set them up
// 2. Abstract as much fuzz logic from the core setup loop, which is already complex enough as is
func (s *KeeperTestSuite) prepareNumPositionSlice(ranges [][]int64, baseNumPositions int, fuzzNumPositions bool) ([]int, int) {
	// Create slice representing number of positions for each range index.
	// Default case is `numPositions` on each range unless fuzzing is turned on.
	numPositionsPerRange := make([]int, len(ranges))
	totalPositions := 0

	// Loop through each range and set number of positions, fuzzing if applicable.
	for i := range ranges {
		numPositionsPerRange[i] = baseNumPositions

		// If applicable, fuzz the number of positions on current range
		if fuzzNumPositions {
			// Fuzzed amount should be between 1 and (2 * numPositions) + 1 (up to 100% fuzz both ways from numPositions)
			numPositionsPerRange[i] = int(fuzzInt64(int64(baseNumPositions), 2))
		}

		// Track total positions
		totalPositions += numPositionsPerRange[i]
	}

	return numPositionsPerRange, totalPositions
}

// executeRandomizedSwap executes a swap against the pool, fuzzing the swap amount if applicable.
// The direction of the swap is chosen randomly, but the swap function used is always SwapInGivenOut to
// ensure it is always possible to swap against the pool without having to use lower level calc functions.
// TODO: Make swaps that target getting to a tick boundary exactly
func (s *KeeperTestSuite) executeRandomizedSwap(pool types.ConcentratedPoolExtension, swapAddresses []sdk.AccAddress, baseSwapAmount osmomath.Int, fuzzSwap bool) (sdk.Coin, sdk.Coin) {
	// Quietly skip if no swap assets or swap addresses provided
	if (baseSwapAmount == osmomath.Int{}) || len(swapAddresses) == 0 {
		return sdk.Coin{}, sdk.Coin{}
	}

	poolLiquidity := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
	s.Require().True(len(poolLiquidity) == 1 || len(poolLiquidity) == 2, "Pool liquidity should be in one or two tokens")

	// Choose swap address
	swapAddressIndex := fuzzInt64(int64(len(swapAddresses)-1), 1)
	swapAddress := swapAddresses[swapAddressIndex]

	// Decide which denom to swap in & out

	var swapInDenom, swapOutDenom string
	if len(poolLiquidity) == 1 {
		// If all pool liquidity is in one token, swap in the other token
		swapOutDenom = poolLiquidity[0].Denom
		if swapOutDenom == pool.GetToken0() {
			swapInDenom = pool.GetToken1()
		} else {
			swapInDenom = pool.GetToken0()
		}
	} else {
		// Otherwise, randomly determine which denom to swap in & out
		swapInDenom, swapOutDenom = randOrder(pool.GetToken0(), pool.GetToken1())
	}

	// TODO: pick a more granular amount to fund without losing ability to swap at really high/low ticks
	swapInFunded := sdk.NewCoin(swapInDenom, osmomath.Int(osmomath.MustNewDecFromStr("10000000000000000000000000000000000000000")))
	s.FundAcc(swapAddress, sdk.NewCoins(swapInFunded))

	baseSwapOutAmount := osmomath.MinInt(baseSwapAmount, poolLiquidity.AmountOf(swapOutDenom).ToLegacyDec().Mul(osmomath.MustNewDecFromStr("0.5")).TruncateInt())
	if fuzzSwap {
		// Fuzz +/- 100% of base swap amount
		baseSwapOutAmount = osmomath.NewInt(fuzzInt64(baseSwapOutAmount.Int64(), 2))
	}

	swapOutCoin := sdk.NewCoin(swapOutDenom, baseSwapOutAmount)

	// If the swap we're about to execute will not generate enough input, we skip the swap.
	if swapOutDenom == pool.GetToken1() {
		pool, err := s.Clk.GetPoolById(s.Ctx, pool.GetId())
		s.Require().NoError(err)

		poolSpotPrice := pool.GetCurrentSqrtPrice().PowerInteger(2)
		minSwapOutAmount := poolSpotPrice.Mul(osmomath.SmallestBigDec()).TruncateDec().Dec().TruncateInt()
		poolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
		if poolBalances.AmountOf(swapOutDenom).LTE(minSwapOutAmount) {
			return sdk.Coin{}, sdk.Coin{}
		}
	}

	// Note that we set the price limit to zero to ensure that the swap can execute in either direction (gets automatically set to correct limit)
	swappedIn, swappedOut, _, err := s.Clk.SwapInAmtGivenOut(s.Ctx, swapAddress, pool, swapOutCoin, swapInDenom, pool.GetSpreadFactor(s.Ctx), osmomath.ZeroBigDec())
	s.Require().NoError(err)

	return swappedIn, swappedOut
}

func randOrder[T any](a, b T) (T, T) {
	if rand.Int()%2 == 0 {
		return a, b
	}
	return b, a
}

// addRandomizedBlockTime adds the given block time to the context, fuzzing the added time if applicable.
func (s *KeeperTestSuite) addRandomizedBlockTime(baseTimeToAdd time.Duration, fuzzTime bool) time.Duration {
	if baseTimeToAdd != time.Duration(0) {
		timeToAdd := baseTimeToAdd
		if fuzzTime {
			// Fuzz +/- 100% of base time to add
			timeToAdd = time.Duration(fuzzInt64(int64(baseTimeToAdd), 2))
		}

		s.AddBlockTime(timeToAdd)
	}

	return baseTimeToAdd
}

// trackEmittedIncentives takes in a cumulative incentives distributed and the last time this number was updated.
// CONTRACT: cumulativeTrackedIncentives has been updated immediately before each new incentive record that was created
func (s *KeeperTestSuite) trackEmittedIncentives(cumulativeTrackedIncentives sdk.DecCoins, lastTrackerUpdateTime time.Time) (sdk.DecCoins, time.Time) {
	// Fetch all incentive records across all pools
	allPools, err := s.Clk.GetPools(s.Ctx)
	s.Require().NoError(err)
	allIncentiveRecords := make([]types.IncentiveRecord, 0)
	for _, pool := range allPools {
		curPoolRecords, err := s.Clk.GetAllIncentiveRecordsForPool(s.Ctx, pool.GetId())
		s.Require().NoError(err)

		allIncentiveRecords = append(allIncentiveRecords, curPoolRecords...)
	}

	// Track new emissions since last checkpoint, factoring in when each incentive record started emitting
	updatedTrackedIncentives := cumulativeTrackedIncentives
	for _, incentiveRecord := range allIncentiveRecords {
		recordStartTime := incentiveRecord.IncentiveRecordBody.StartTime

		// If the record hasn't started emitting yet, skip it
		if recordStartTime.After(s.Ctx.BlockTime()) {
			continue
		}

		secondsEmitted := osmomath.ZeroDec()
		if recordStartTime.Before(lastTrackerUpdateTime) {
			// If the record started emitting prior to the last incentiveCreationTime (the last time we checkpointed),
			// then we assume it has been emitting for the whole period since then.
			secondsEmitted = osmomath.NewDec(int64(s.Ctx.BlockTime().Sub(lastTrackerUpdateTime))).QuoInt64(int64(time.Second))
		} else if recordStartTime.Before(s.Ctx.BlockTime()) {
			// If the record started emitting between the last incentiveCreationTime and now, then we only track the
			// emissions between when it started and now.
			secondsEmitted = osmomath.NewDec(int64(s.Ctx.BlockTime().Sub(recordStartTime))).QuoInt64(int64(time.Second))
		}

		emissionRate := incentiveRecord.IncentiveRecordBody.EmissionRate
		incentiveDenom := incentiveRecord.IncentiveRecordBody.RemainingCoin.Denom

		// Track emissions for the current record
		emittedAmount := emissionRate.Mul(secondsEmitted)
		emittedDecCoin := sdk.NewDecCoinFromDec(incentiveDenom, emittedAmount)
		updatedTrackedIncentives = updatedTrackedIncentives.Add(emittedDecCoin)
	}

	return updatedTrackedIncentives, s.Ctx.BlockTime()
}

// getInitialPositionAssets returns the assets required for the first position in a pool to set the initial current tick to the given value.
func (s *KeeperTestSuite) getInitialPositionAssets(pool types.ConcentratedPoolExtension, initialCurrentTick int64) sdk.Coins {
	requiredPrice, err := math.TickToPrice(initialCurrentTick)
	s.Require().NoError(err)

	// Calculate asset amounts that would be required to get the required spot price (rounding up on asset1 to ensure we stay in the intended tick)
	asset0Amount := osmomath.NewInt(100000000000000)
	asset1Amount := osmomath.BigDecFromDec(osmomath.NewDecFromInt(asset0Amount)).Mul(requiredPrice).Ceil().Dec().TruncateInt()

	assetCoins := sdk.NewCoins(
		sdk.NewCoin(pool.GetToken0(), asset0Amount),
		sdk.NewCoin(pool.GetToken1(), asset1Amount),
	)

	return assetCoins
}

// getFuzzedAssets returns the base asset amount, fuzzing each asset if applicable
func getRandomizedAssets(baseAssets sdk.Coins, fuzzAssets bool) sdk.Coins {
	finalAssets := baseAssets
	if fuzzAssets {
		fuzzedAssets := make([]sdk.Coin, len(baseAssets))
		for coinIndex, coin := range baseAssets {
			// Fuzz +/- 100% of current amount
			newAmount := fuzzInt64(coin.Amount.Int64(), 2)
			fuzzedAssets[coinIndex] = sdk.NewCoin(coin.Denom, osmomath.NewInt(newAmount))
		}

		finalAssets = fuzzedAssets
	}

	return finalAssets
}

// fuzzInt64 fuzzes an int64 number uniformly within a range defined by `multiplier` and centered on the provided `intToFuzz`.
func fuzzInt64(intToFuzz int64, multiplier int64) int64 {
	return (rand.Int63() % (multiplier * intToFuzz)) + 1
}
