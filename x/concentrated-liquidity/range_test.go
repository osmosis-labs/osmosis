package concentrated_liquidity_test

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/app/apptesting"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

type RangeTestParams struct {
	// -- Base amounts --

	// Base number of assets for each position
	baseAssets sdk.Coins
	// Base number of positions for each range
	baseNumPositions int
	// Base amount to swap for each swap
	baseSwapAmount sdk.Int
	// Base amount to add after each new position
	baseTimeBetweenJoins time.Duration
	// Base incentive records to have on pool
	baseIncentiveRecords []types.IncentiveRecord
	// List of addresses to swap from (randomly selected for each swap)
	numSwapAddresses int

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
}

var (
	DefaultRangeTestParams = RangeTestParams{
		// Base amounts
		baseNumPositions:     10,
		baseAssets:           sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(5000000000)), sdk.NewCoin(USDC, sdk.NewInt(5000000000))),
		baseTimeBetweenJoins: time.Hour,
		baseSwapAmount:       sdk.NewInt(10000000),
		numSwapAddresses:     10,

		// Fuzz params
		fuzzNumPositions:     true,
		fuzzAssets:           true,
		fuzzSwapAmounts:      true,
		fuzzTimeBetweenJoins: true,
	}
)

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

	// TODO: support incentive fuzzing (use to `totalTimeElapsed` to track emitted amounts)

	// --- Position setup ---

	// This loop runs through each given tick range and does the following at each iteration:
	// 1. Set up a position
	// 2. Let time elapse
	// 3. Execute a swap
	totalLiquidity, totalAssets, totalTimeElapsed, allPositionIds, lastVisitedBlockIndex := sdk.ZeroDec(), sdk.NewCoins(), time.Duration(0), []uint64{}, 0
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
			roundingError := sdk.NewCoins(sdk.NewCoin(pool.GetToken0(), sdk.OneInt()), sdk.NewCoin(pool.GetToken1(), sdk.OneInt()))
			s.FundAcc(curAddr, curAssets.Add(roundingError...))

			// Set up position
			curPositionId, actualAmt0, actualAmt1, curLiquidity, actualLowerTick, actualUpperTick, err := s.clk.CreatePosition(s.Ctx, pool.GetId(), curAddr, curAssets, sdk.ZeroInt(), sdk.ZeroInt(), ranges[curRange][0], ranges[curRange][1])
			s.Require().NoError(err)

			// Ensure position was set up correctly and didn't break global invariants
			s.Require().Equal(ranges[curRange][0], actualLowerTick)
			s.Require().Equal(ranges[curRange][1], actualUpperTick)
			s.assertGlobalInvariants()

			// Let time elapse after join if applicable
			timeElapsed := s.addRandomizedBlockTime(testParams.baseTimeBetweenJoins, testParams.fuzzTimeBetweenJoins)

			// Execute swap against pool if applicable
			swappedIn, swappedOut := s.executeRandomizedSwap(pool, swapAddresses, testParams.baseSwapAmount, testParams.fuzzSwapAmounts)
			s.assertGlobalInvariants()

			// Track changes to state
			actualAddedCoins := sdk.NewCoins(sdk.NewCoin(pool.GetToken0(), actualAmt0), sdk.NewCoin(pool.GetToken1(), actualAmt1))
			totalAssets = totalAssets.Add(actualAddedCoins...).Add(swappedIn).Sub(sdk.NewCoins(swappedOut))
			totalLiquidity = totalLiquidity.Add(curLiquidity)
			totalTimeElapsed = totalTimeElapsed + timeElapsed
			allPositionIds = append(allPositionIds, curPositionId)
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
func (s *KeeperTestSuite) executeRandomizedSwap(pool types.ConcentratedPoolExtension, swapAddresses []sdk.AccAddress, baseSwapAmount sdk.Int, fuzzSwap bool) (sdk.Coin, sdk.Coin) {
	// Quietly skip if no swap assets or swap addresses provided
	if (baseSwapAmount == sdk.Int{}) || len(swapAddresses) == 0 {
		return sdk.Coin{}, sdk.Coin{}
	}

	binaryFlip := rand.Int() % 2
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
		if binaryFlip == 0 {
			swapInDenom = pool.GetToken0()
			swapOutDenom = pool.GetToken1()
		} else {
			swapInDenom = pool.GetToken1()
			swapOutDenom = pool.GetToken0()
		}
	}

	// TODO: pick a more granular amount to fund without losing ability to swap at really high/low ticks
	swapInFunded := sdk.NewCoin(swapInDenom, sdk.Int(sdk.MustNewDecFromStr("10000000000000000000000000000000000000000")))
	s.FundAcc(swapAddress, sdk.NewCoins(swapInFunded))

	baseSwapOutAmount := sdk.MinInt(baseSwapAmount, poolLiquidity.AmountOf(swapOutDenom).ToDec().Mul(sdk.MustNewDecFromStr("0.5")).TruncateInt())
	if fuzzSwap {
		// Fuzz +/- 100% of base swap amount
		baseSwapOutAmount = sdk.NewInt(fuzzInt64(baseSwapOutAmount.Int64(), 2))
	}

	swapOutCoin := sdk.NewCoin(swapOutDenom, baseSwapOutAmount)

	// Note that we set the price limit to zero to ensure that the swap can execute in either direction (gets automatically set to correct limit)
	swappedIn, swappedOut, _, _, _, err := s.clk.SwapInAmtGivenOut(s.Ctx, swapAddress, pool, swapOutCoin, swapInDenom, pool.GetSpreadFactor(s.Ctx), sdk.ZeroDec())
	s.Require().NoError(err)

	return swappedIn, swappedOut
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

// getFuzzedAssets returns the base asset amount, fuzzing each asset if applicable
func getRandomizedAssets(baseAssets sdk.Coins, fuzzAssets bool) sdk.Coins {
	finalAssets := baseAssets
	if fuzzAssets {
		fuzzedAssets := make([]sdk.Coin, len(baseAssets))
		for coinIndex, coin := range baseAssets {
			// Fuzz +/- 100% of current amount
			newAmount := fuzzInt64(coin.Amount.Int64(), 2)
			fuzzedAssets[coinIndex] = sdk.NewCoin(coin.Denom, sdk.NewInt(newAmount))
		}

		finalAssets = fuzzedAssets
	}

	return finalAssets
}

// fuzzInt64 fuzzes an int64 number uniformly within a range defined by `multiplier` and centered on the provided `intToFuzz`.
func fuzzInt64(intToFuzz int64, multiplier int64) int64 {
	return (rand.Int63() % (multiplier * intToFuzz)) + 1
}
