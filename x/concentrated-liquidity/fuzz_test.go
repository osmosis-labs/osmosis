package concentrated_liquidity_test

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

const (
	maxAmountDeposited  = 999_999_999_999_999_999
	initialNumPositions = 20

	defaultNumSwaps     = 30
	defaultNumPositions = 10
)

func (s *KeeperTestSuite) TestFuzz() {
	s.FuzzTest(30, 10, 100)
}

func (s *KeeperTestSuite) TestFuzz_GivenSeed() {
	r := rand.New(rand.NewSource(1688513262))
	s.individualFuzz(r, 0, defaultNumSwaps, defaultNumPositions)
}

// pre-condition: poolId exists, and has at least one position
func (s *KeeperTestSuite) FuzzTest(numSwaps int, numPositions int, numFuzzes int) {
	seed := time.Now().Unix()

	wg := &sync.WaitGroup{}

	for i := 0; i < numFuzzes; i++ {
		i := i
		wg.Add(1)

		currentSeed := seed + int64(i)
		r := rand.New(rand.NewSource(currentSeed))

		currentSuite := &KeeperTestSuite{}
		currentSuite.SetS(s)
		currentSuite.SetT(s.T())

		currentSuite.Run(fmt.Sprintf("Fuzz %d, seed: %d", i, currentSeed), func() {
			currentSuite.T().Parallel()

			currentSuite.individualFuzz(r, i, numSwaps, numPositions)
			wg.Done()
		})
	}

	// wg.Wait()
}

func (s *KeeperTestSuite) individualFuzz(r *rand.Rand, fuzzNum int, numSwaps int, numPositions int) {
	s.SetupTest()

	spreadFactors := types.DefaultParams().AuthorizedSpreadFactors
	numSpreadFactors := len(spreadFactors)

	spreadFactor := spreadFactors[r.Intn(numSpreadFactors)]
	pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, spreadFactor)

	initialAmt0 := randomIntAmount(r)
	initialAmt1 := randomIntAmount(r)

	defaultCoins := sdk.NewCoins(sdk.NewCoin(ETH, initialAmt0), sdk.NewCoin(USDC, initialAmt1))
	s.CreateFullRangePosition(pool, defaultCoins)

	// Refetch pool
	pool, err := s.clk.GetPoolById(s.Ctx, pool.GetId())
	s.Require().NoError(err)

	fmt.Printf("SINGLE FUZZ START: %d. initialAmt0 %s initialAmt1 %s \n", fuzzNum, initialAmt0, initialAmt1)

	s.fuzzTestWithSeed(r, pool.GetId(), numSwaps, numPositions)

	// validate if any errrs

	for _, err := range s.collectedErrors {
		fmt.Println("collected error: ", err)
	}
	// If error observed, note seed value and rerun with TestFuzz_GivenSeed
	// search for "swap error:" in logs for details.
	s.Require().Empty(s.collectedErrors)
}

type fuzzState struct {
	r      *rand.Rand
	poolId int
}

func (s *KeeperTestSuite) fuzzTestWithSeed(r *rand.Rand, poolId uint64, numSwaps int, numPositions int) {
	// Add 1000 random positions
	for i := 0; i < initialNumPositions; i++ {
		s.addRandomPositonMinMaxOneSpacing(r, poolId)
	}
	s.assertWithdrawAllInvariant()

	fmt.Printf("\n\n\n--------------------Positions Pre-Created Beginning Fuzz--------------------\n\n\n")

	// Fuzz by swapping and adding/removing liquidity in-between

	completedSwaps := 0
	completedPositions := 0
	targetActions := numSwaps + numPositions
	for i := 0; i < targetActions; i++ {
		fmt.Printf("\n\n\n>>>>>>>>>>>>>> Begin Action\n")

		doSwap := s.selectAction(r, numSwaps, numPositions, completedSwaps, completedPositions)
		if doSwap {
			// s.swap(r)
			s.randomSwap(r, poolId)
			completedSwaps++
		} else {
			s.addOrRemoveLiquidity(r, poolId)
			completedPositions++
		}

		s.assertWithdrawAllInvariant()
	}
}

func (s *KeeperTestSuite) randomSwap(r *rand.Rand, poolId uint64) {

	for didSwap := false; !didSwap; {

		pool, err := s.clk.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		zfo := s.chooseSwapDirection(r, pool)

		// High level decision, decide which swap strategy to do.
		// 1. Swap a random amount
		// 2. Swap near next tick boundary
		// 3. Swap to later tick boundary (TODO)
		swapStrategy := r.Intn(3)
		if swapStrategy == 0 {
			didSwap = s.swapRandomAmount(r, pool, zfo)
		} else if swapStrategy == 1 {
			didSwap = s.swapNearNextTickBoundary(r, pool, zfo)
		} else {
			didSwap = s.swapNearInitializedTickBoundary(r, pool, zfo)
		}

		if !didSwap {
			fmt.Printf("swap failed for acceptable reasons, retrying \n\n")
		}
	}
}

func (s *KeeperTestSuite) swapRandomAmount(r *rand.Rand, pool types.ConcentratedPoolExtension, zfo bool) (didSwap bool) {
	fmt.Println("swap type: random amount")
	swapInDenom, swapOutDenom := zfoToDenoms(zfo, pool)
	swapAmt := randomIntAmount(r)
	swapInCoin := sdk.NewCoin(swapInDenom, swapAmt)
	return s.swap(pool, swapInCoin, swapOutDenom)
}

func (s *KeeperTestSuite) swapNearNextTickBoundary(r *rand.Rand, pool types.ConcentratedPoolExtension, zfo bool) (didSwap bool) {
	fmt.Println("swap type: near next tick boundary")
	targetTick := pool.GetCurrentTick()
	if zfo {
		targetTick -= 1
	} else {
		targetTick += 1
	}
	return s.swapNearTickBoundary(r, pool, targetTick, zfo)
}

func (s *KeeperTestSuite) swapNearInitializedTickBoundary(r *rand.Rand, pool types.ConcentratedPoolExtension, zfo bool) (didSwap bool) {
	fmt.Println("swap type: near initialized tick boundary")

	ss := swapstrategy.New(zfo, sdk.ZeroDec(), s.App.GetKey(types.ModuleName), sdk.ZeroDec())

	iter := ss.InitializeNextTickIterator(s.Ctx, pool.GetId(), pool.GetCurrentTick())
	defer iter.Close()

	if !iter.Valid() {
		return false
	}

	s.Require().True(iter.Valid())

	nextInitializedTick, err := types.TickIndexFromBytes(iter.Key())
	s.Require().NoError(err)

	return s.swapNearTickBoundary(r, pool, nextInitializedTick, zfo)
}

func (s *KeeperTestSuite) swapNearTickBoundary(r *rand.Rand, pool types.ConcentratedPoolExtension, targetTick int64, zfo bool) (didSwap bool) {
	swapInDenom, swapOutDenom := zfoToDenoms(zfo, pool)
	// TODO: Confirm accuracy of this method.
	amountInRequired, curLiquidity, _ := s.computeSwapAmounts(pool.GetId(), pool.GetCurrentSqrtPrice(), targetTick, zfo, false)

	// Decide if below, exactly, or above target tick

	poolSpotPrice := pool.GetCurrentSqrtPrice().Power(osmomath.NewBigDec(2))
	fmt.Printf("pool: tick %d, spot price: %s, liq %s \n", pool.GetCurrentTick(), poolSpotPrice, curLiquidity)

	amountInRequired = tickAmtChange(r, amountInRequired)

	swapInCoin := sdk.NewCoin(swapInDenom, amountInRequired.TruncateInt())
	return s.swap(pool, swapInCoin, swapOutDenom)
}

// change tick amount to be at, above or below the target amount
func tickAmtChange(r *rand.Rand, targetAmount sdk.Dec) sdk.Dec {
	changeType := r.Intn(3)

	// Generate a random percentage under 0.1%
	randChangePercent := sdk.NewDec(rand.Int63n(1)).QuoInt64(1000)
	change := targetAmount.Mul(randChangePercent)

	change = sdk.MaxDec(sdk.NewDec(1), randChangePercent)

	switch changeType {
	case 0:
		fmt.Printf("tick amt change type: no change, orig: %s \n", targetAmount)
		// do nothing
		return targetAmount
	case 1:
		// above tick
		change = change.Ceil()
		fmt.Printf("tick amt change type: beyond tick, orig: %s  change added %s\n", targetAmount, change)
		return targetAmount.Add(change.TruncateDec())
	}

	if targetAmount.LTE(sdk.OneDec()) {
		fmt.Printf("tick amt change type: no change, orig: %s \n", targetAmount)
		return targetAmount
	}

	// below tick
	change = change.TruncateDec()
	fmt.Printf("tick amt change type: not reaching tick, orig: %s change subtracted: %s\n", targetAmount, change)
	return targetAmount.Sub(change.TruncateDec())
}

func (s *KeeperTestSuite) swap(pool types.ConcentratedPoolExtension, swapInFunded sdk.Coin, swapOutDenom string) (didSwap bool) {
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(swapInFunded))
	// // Execute swap
	fmt.Printf("swap in: %s\n", swapInFunded)
	cacheCtx, write := s.Ctx.CacheContext()
	_, _, _, err := s.clk.SwapOutAmtGivenIn(cacheCtx, s.TestAccs[0], pool, swapInFunded, swapOutDenom, pool.GetSpreadFactor(s.Ctx), sdk.ZeroDec())
	if errors.As(err, &types.InvalidAmountCalculatedError{}) {
		// If the swap we're about to execute will not generate enough output, we skip the swap.
		// it would error for a real user though. This is good though, since that user would just be burning funds.
		if err.(types.InvalidAmountCalculatedError).Amount.IsZero() {
			return false
		}
	}
	if err != nil {
		fmt.Printf("swap error: %s\n", err.Error())
		// Add error to list of errors. Will fail at the end of the fuzz run in hih level test.
		s.collectedErrors = append(s.collectedErrors, err)
		return false
	}

	// Write only if no error
	write()

	return true
}

func (s *KeeperTestSuite) chooseSwapDirection(r *rand.Rand, pool types.ConcentratedPoolExtension) (zfo bool) {
	poolLiquidity := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
	s.Require().True(len(poolLiquidity) == 1 || len(poolLiquidity) == 2, "Pool liquidity should be in one or two tokens")

	if len(poolLiquidity) == 1 {
		// If all pool liquidity is in one token, swap in the other token
		swapOutDenom := poolLiquidity[0].Denom
		if swapOutDenom == pool.GetToken0() {
			return false
		} else {
			return true
		}
	}
	return r.Int()%2 == 0
}

func zfoToDenoms(zfo bool, pool types.ConcentratedPoolExtension) (swapInDenom, swapOutDenom string) {
	if zfo {
		return pool.GetToken0(), pool.GetToken1()
	} else {
		return pool.GetToken1(), pool.GetToken0()
	}
}

// if true swap, if false, LP
func (s *KeeperTestSuite) selectAction(r *rand.Rand, numSwaps, numPositions, completedSwaps, completedPositions int) bool {
	if completedSwaps == numSwaps {
		return false
	}
	if completedPositions == numPositions {
		return true
	}

	if numPositions == 0 {
		return false
	}

	return r.Intn(2) == 0
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
// Add or remove liquidity

func (s *KeeperTestSuite) addOrRemoveLiquidity(r *rand.Rand, poolId uint64) {

	// shouldAddPosition := s.selectAddOrRemove(r)

	if true {
		s.addRandomPositonMinMaxOneSpacing(r, poolId)
	} else {
		fmt.Println("removing position")
		// s.removeLiquidity(r, randomizedAssets)
	}

}

// if true add position, if false remove position
func (s *KeeperTestSuite) selectAddOrRemove(r *rand.Rand) bool {
	if len(s.positionIds) == 0 {
		return true
	}
	return r.Intn(2) == 0
}

func (s *KeeperTestSuite) addRandomPositonMinMaxOneSpacing(r *rand.Rand, poolId uint64) {
	s.addRandomPositon(r, poolId, types.MinInitializedTick, types.MaxTick, 1)
}

func (s *KeeperTestSuite) addRandomPositon(r *rand.Rand, poolId uint64, minTick, maxTick int64, tickSpacing int64) {
	tokenDesired0 := sdk.NewCoin(ETH, sdk.NewInt(rand.Int63n(maxAmountDeposited)))
	tokenDesired1 := sdk.NewCoin(USDC, sdk.NewInt(rand.Int63n(maxAmountDeposited)))
	tokensDesired := sdk.NewCoins(tokenDesired0, tokenDesired1)

	s.FundAcc(s.TestAccs[0], tokensDesired)

	lowerTick := roundTickDownSpacing(rand.Int63n(maxTick-minTick+1)+minTick, tickSpacing)
	// lowerTick <= upperTick <= maxTick
	upperTick := roundTickDownSpacing(maxTick-rand.Int63n(int64(math.Abs(float64(maxTick-lowerTick)))), tickSpacing)

	fmt.Println("creating position: ", "accountName", "lowerTick", lowerTick, "upperTick", upperTick, "token0Desired", tokenDesired0, "tokenDesired1", tokenDesired1)

	positionId, amt0, amt1, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, s.TestAccs[0], tokensDesired, sdk.ZeroInt(), sdk.ZeroInt(), types.MinInitializedTick, types.MaxTick)
	s.Require().NoError(err)
	fmt.Printf("actually created: %s%s %s%s \n", amt0, ETH, amt1, USDC)

	s.positionIds = append(s.positionIds, positionId)
}

func roundTickDownSpacing(tickIndex int64, tickSpacing int64) int64 {
	// Round the tick index down to the nearest tick spacing if the tickIndex is in between authorized tick values
	// Note that this is Euclidean modulus.
	// The difference from default Go modulus is that Go default results
	// in a negative remainder when the dividend is negative.
	// Consider example tickIndex = -17, tickSpacing = 10
	// tickIndexModulus = tickIndex % tickSpacing = -7
	// tickIndexModulus = -7 + 10 = 3
	// tickIndex = -17 - 3 = -20
	tickIndexModulus := tickIndex % tickSpacing
	if tickIndexModulus < 0 {
		tickIndexModulus += tickSpacing
	}

	if tickIndexModulus != 0 {
		tickIndex = tickIndex - tickIndexModulus
	}
	return tickIndex
}

func randomIntAmount(r *rand.Rand) sdk.Int {
	return sdk.NewInt(rand.Int63n(maxAmountDeposited))
}
