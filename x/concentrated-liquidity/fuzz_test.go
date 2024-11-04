//go:build !norace

package concentrated_liquidity_test

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

const (
	maxAmountDeposited  = 999_999_999_999_999_999
	initialNumPositions = 20

	defaultNumSwaps     = 30
	defaultNumPositions = 10
)

type swapAmountsMismatchErr struct {
	swapInFunded       sdk.Coin
	amountInSwapResult sdk.Coin
	diff               osmomath.Int
}

func (e swapAmountsMismatchErr) Error() string {
	return fmt.Sprintf("amounts in mismatch, original %s, swapped in given out: %s, difference of %s", e.swapInFunded, e.amountInSwapResult, e.diff)
}

func TestFuzz_Many(t *testing.T) {
	fuzz(t, defaultNumSwaps, defaultNumPositions, 10)
}

func (s *KeeperTestSuite) TestFuzz_GivenSeed() {
	// Seed 1688572291 - gives mismatch between tokenIn given to "out given in" and token in returned from "in given out"
	// Seed 1688658883- causes an error in swap in given out due to rounding (acceptable).
	r := rand.New(rand.NewSource(1688658883))
	s.individualFuzz(r, 0, 30, 10)

	s.validateNoErrors(s.collectedErrors)
}

// pre-condition: poolId exists, and has at least one position
func fuzz(t *testing.T, numSwaps int, numPositions int, numIterations int) {
	seed := time.Now().Unix()

	for i := 0; i < numIterations; i++ {
		i := i

		currentSeed := seed + int64(i)
		r := rand.New(rand.NewSource(currentSeed))

		currentSuite := &KeeperTestSuite{}
		currentSuite.SetT(t)
		currentSuite.seed = currentSeed
		currentSuite.iteration = i

		t.Run(fmt.Sprintf("Fuzz %d, seed: %d", i, currentSeed), func(t *testing.T) {
			// This is commented out temporarily to avoid issues with the wasmStoreKey being
			// a global in the wasm light client.
			// See https://github.com/cosmos/ibc-go/blob/modules/light-clients/08-wasm/v0.1.0%2Bibc-go-v7.3-wasmvm-v1.5/modules/light-clients/08-wasm/internal/ibcwasm/wasm.go#L15-L17
			// Once we move to ibc v8 we can make these tests run in parallel again.
			//t.Parallel()

			currentSuite.individualFuzz(r, i, numSwaps, numPositions)
		})
	}
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
	pool, err := s.Clk.GetPoolById(s.Ctx, pool.GetId())
	s.Require().NoError(err)

	fmt.Printf("SINGLE FUZZ START: %d. initialAmt0 %s initialAmt1 %s \n", fuzzNum, initialAmt0, initialAmt1)

	s.fuzzTestWithSeed(r, pool.GetId(), numSwaps, numPositions)

	s.validateNoErrors(s.collectedErrors)
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
			fatalErr := s.randomSwap(r, poolId)
			completedSwaps++
			if fatalErr {
				fmt.Println("Fatal error, exiting")
				return
			}
		} else {
			s.addOrRemoveLiquidity(r, poolId)
			completedPositions++
		}

		if r.Intn(2) == 0 {
			// at some interval, transfer position to a new account
			s.transferRandomPosition(r)
		}

		s.assertGlobalInvariants(ExpectedGlobalRewardValues{})
	}
}

func (s *KeeperTestSuite) randomSwap(r *rand.Rand, poolId uint64) (fatalErr bool) {
	pool, err := s.Clk.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	updateStrategy := func() (swapStrategy int, zfo bool) {
		zfo = s.chooseSwapDirection(r, pool)

		// High level decision, decide which swap strategy to do.
		// 1. Swap a random amount
		// 2. Swap near next tick boundary
		// 3. Swap to later tick boundary (TODO)
		swapStrategy = r.Intn(3)

		return swapStrategy, zfo
	}

	swapStrategy, zfo := updateStrategy()

	for didSwap := false; !didSwap; {

		if swapStrategy == 0 {
			didSwap, fatalErr = s.swapRandomAmount(r, pool, zfo)
		} else if swapStrategy == 1 {
			didSwap, fatalErr = s.swapNearNextTickBoundary(r, pool, zfo)
		} else {
			didSwap, fatalErr = s.swapNearInitializedTickBoundary(r, pool, zfo)
		}

		if fatalErr {
			return true
		}
		if !didSwap {
			fmt.Printf("swap failed for acceptable reasons, retrying \n\n")
		}

		// Only update strategy if previous one succeeded to prevent accidental skip
		// of certain strategies.
		swapStrategy, zfo = updateStrategy()
	}
	return false
}

func (s *KeeperTestSuite) swapRandomAmount(r *rand.Rand, pool types.ConcentratedPoolExtension, zfo bool) (didSwap bool, fatalErr bool) {
	fmt.Println("swap type: random amount")
	swapInDenom, swapOutDenom := zfoToDenoms(zfo, pool)
	swapAmt := randomIntAmount(r)
	swapInCoin := sdk.NewCoin(swapInDenom, swapAmt)
	return s.swap(pool, swapInCoin, swapOutDenom)
}

func (s *KeeperTestSuite) swapNearNextTickBoundary(r *rand.Rand, pool types.ConcentratedPoolExtension, zfo bool) (didSwap bool, fatalErr bool) {
	fmt.Println("swap type: near next tick boundary")
	targetTick := pool.GetCurrentTick()
	if zfo {
		targetTick -= 1
	} else {
		targetTick += 1
	}
	// TODO: remove this limit upon completion of the refactor in:
	// https://github.com/osmosis-labs/osmosis/issues/5726
	// Due to an intermediary refactor step where we have
	// full range positions created in the extended full range it
	// sometimes tries to swap to the V2 MinInitializedTick that
	// is not supported yet by the rest of the system.
	if targetTick < types.MinInitializedTick {
		return false, false
	}
	return s.swapNearTickBoundary(r, pool, targetTick, zfo)
}

func (s *KeeperTestSuite) swapNearInitializedTickBoundary(r *rand.Rand, pool types.ConcentratedPoolExtension, zfo bool) (didSwap bool, fatalErr bool) {
	fmt.Println("swap type: near initialized tick boundary")

	ss := swapstrategy.New(zfo, osmomath.ZeroBigDec(), s.App.GetKey(types.ModuleName), osmomath.ZeroDec())

	iter := ss.InitializeNextTickIterator(s.Ctx, pool.GetId(), pool.GetCurrentTick())
	defer iter.Close()

	if !iter.Valid() {
		return false, false
	}

	s.Require().True(iter.Valid())

	nextInitializedTick, err := types.TickIndexFromBytes(iter.Key())
	s.Require().NoError(err)

	return s.swapNearTickBoundary(r, pool, nextInitializedTick, zfo)
}

func (s *KeeperTestSuite) swapNearTickBoundary(r *rand.Rand, pool types.ConcentratedPoolExtension, targetTick int64, zfo bool) (didSwap bool, fatalErr bool) {
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
func tickAmtChange(r *rand.Rand, targetAmount osmomath.Dec) osmomath.Dec {
	changeType := r.Intn(3)

	// Generate a random percentage under 0.1%
	randChangePercent := osmomath.NewDec(r.Int63n(1)).QuoInt64(1000)
	change := targetAmount.Mul(randChangePercent)

	change = osmomath.MaxDec(osmomath.NewDec(1), randChangePercent)

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

	if targetAmount.LTE(osmomath.OneDec()) {
		fmt.Printf("tick amt change type: no change, orig: %s \n", targetAmount)
		return targetAmount
	}

	// below tick
	change = change.TruncateDec()
	fmt.Printf("tick amt change type: not reaching tick, orig: %s change subtracted: %s\n", targetAmount, change)
	return targetAmount.Sub(change.TruncateDec())
}

func (s *KeeperTestSuite) swap(pool types.ConcentratedPoolExtension, swapInFunded sdk.Coin, swapOutDenom string) (didSwap bool, fatalErr bool) {
	// Reason for adding one int:
	// Seed 1688658883- causes an error in swap in given out due to rounding (acceptable). This is because we use
	// token out from "swap out given in" as an input to "in given out". "in given out" rounds by one in pool's favor
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(swapInFunded).Add(sdk.NewCoin(swapInFunded.Denom, osmomath.OneInt())))
	// // Execute swap
	fmt.Printf("swap in: %s\n", swapInFunded)
	cacheCtx, writeOutGivenIn := s.Ctx.CacheContext()
	_, tokenOut, _, err := s.Clk.SwapOutAmtGivenIn(cacheCtx, s.TestAccs[0], pool, swapInFunded, swapOutDenom, pool.GetSpreadFactor(s.Ctx), osmomath.ZeroBigDec())
	if errors.As(err, &types.InvalidAmountCalculatedError{}) {
		// If the swap we're about to execute will not generate enough output, we skip the swap.
		// it would error for a real user though. This is good though, since that user would just be burning funds.
		if err.(types.InvalidAmountCalculatedError).Amount.IsZero() {
			return false, false
		}
	}
	if err != nil {
		fmt.Printf("swap error in out given in: %s\n", err.Error())
		// Add error to list of errors. Will fail at the end of the fuzz run in high level test.
		s.collectedErrors = append(s.collectedErrors, err)
		return false, false
	}

	// Now, swap in given out with the amount out given by previous swap
	// We expect the returned amountIn to be roughly equal to the original swapInFunded.
	cacheCtx, _ = s.Ctx.CacheContext()
	fmt.Printf("swap out: %s\n", tokenOut)
	amountInSwapResult, _, _, err := s.Clk.SwapInAmtGivenOut(cacheCtx, s.TestAccs[0], pool, tokenOut, swapInFunded.Denom, pool.GetSpreadFactor(s.Ctx), osmomath.ZeroBigDec())
	if errors.As(err, &types.InvalidAmountCalculatedError{}) {
		// If the swap we're about to execute will not generate enough output, we skip the swap.
		// it would error for a real user though. This is good though, since that user would just be burning funds.
		if err.(types.InvalidAmountCalculatedError).Amount.IsZero() {
			return false, false
		}
	}

	if err != nil {
		fmt.Printf("swap error in in given out: %s\n", err.Error())
		// Add error to list of errors. Will fail at the end of the fuzz run in high level test.
		s.collectedErrors = append(s.collectedErrors, err)
		return false, false
	}

	errTolerance := osmomath.ErrTolerance{
		// 2% tolerance
		MultiplicativeTolerance: osmomath.NewDecWithPrec(2, 2),
		// Expected amount in returned from swap "in given out" to be smaller
		// than original amount in given to "out given in".
		// Reason: rounding in pool's favor.
		RoundingDir: osmomath.RoundDown,
	}

	result := errTolerance.CompareBigDec(osmomath.BigDecFromDecMut(swapInFunded.Amount.ToLegacyDec()), osmomath.BigDecFromDecMut(amountInSwapResult.Amount.ToLegacyDec()))

	if result != 0 {
		// Note: did some investigations into why this happens.
		// Seed: 1688572291
		//
		// Logs & Prints (only relevant parts for brevity):
		//
		//
		// swap out given in
		//
		// swap in: 53154819938620036019618231426080065450usdc
		// start sqrt price 1.925114640286395175000000000000000000
		// reached sqrt price 9990000000000000000.001925114640286395490901354505635230
		// liquidity 5315481993862003602.985114363264252712
		// amountIn 53101665118681415983598613194653985385.000000000000000000
		//
		// swap in given out
		// start sqrt price 1.925114640286395175000000000000000000
		// reached sqrt price 5504536865264953043.262903126972924283156861610109232553
		// liquidity 5315481993862003602.985114363264252712
		// amountIn 29259266591865455675844375866509084121.000000000000000000
		//
		// Note that reached sqrt price is different in both cases leading to different amounts in.
		// Traced the values with clmath.py.
		// Conclusion: the small rounding difference under 1 unit leads to such a large difference because
		// the affected sqrt price range is long
		//
		// This is what we get in in given out calculation of sqrt price with non-rounded tokenOut:
		// get_next_sqrt_price_from_amount0_out_round_up(liquidity, sqrtPriceCurrent, amountOutGiven)
		// Decimal('9989999999999999983.247487166393337205203511259650091832')
		//
		// This is what we get in in given out calculation of sqrt price with rounded tokenOut:
		// get_next_sqrt_price_from_amount0_out_round_up(liquidity, sqrtPriceCurrent, amountOutTests)
		// Decimal('5504536865264953043.262903126972924283156861610109232553')
		//
		// This proves that this is a test setup error, not a swap logic error. We need smarter detection of when
		// a small difference between non-rounded tokenOut in swap out given in and the returned tokenOut here leads
		// to a large difference in sqrt price (TBD later).
		s.collectedErrors = append(s.collectedErrors, swapAmountsMismatchErr{swapInFunded: swapInFunded, amountInSwapResult: amountInSwapResult, diff: swapInFunded.Amount.Sub(amountInSwapResult.Amount)})
		return true, false
	}

	// Write out given in only if no error. In given out state is dropped.
	writeOutGivenIn()

	return true, false
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

// validate if any errrs
func (s *KeeperTestSuite) validateNoErrors(possibleErrors []error) {
	fullMsg := ""
	shouldFail := false
	for _, err := range possibleErrors {

		// TODO: figure out if this is OK
		// Answer: Ok for now, due to outofbounds=True restriction
		// Should sanity check that our fuzzer isn't hitting this too often though, that could hit at
		// ineffective fuzz range choice.
		if errors.Is(err, types.SqrtPriceToTickError{OutOfBounds: true}) {
			continue
		}

		// This is expected
		if errors.As(err, &types.ComputedSqrtPriceInequalityError{}) {
			continue
		}
		// TODO: Need to understand why this is happening
		if errors.As(err, &types.OverChargeSwapOutGivenInError{}) {
			continue
		}

		// This is acceptable. See where this error is returned for explanation.
		if errors.As(err, &swapAmountsMismatchErr{}) {
			continue
		}

		shouldFail = true

		msg := fmt.Sprintf("%s\n", err.Error())
		fmt.Println(msg)
		fullMsg += msg
	}

	if shouldFail {
		s.Fail(fmt.Sprintf("failed validation for errors seed: %d iteration: %d, %s", s.seed, s.iteration, fullMsg), fullMsg)
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
	if s.selectAddOrRemove(r) {
		s.addRandomPositonMinMaxOneSpacing(r, poolId)
	} else {
		s.removeRandomPosition(r)
	}

}

// if true add position, if false remove position
func (s *KeeperTestSuite) selectAddOrRemove(r *rand.Rand) bool {
	if len(s.positionData) == 0 {
		return true
	}
	return r.Intn(2) == 0
}

func (s *KeeperTestSuite) addRandomPositonMinMaxOneSpacing(r *rand.Rand, poolId uint64) {
	s.addRandomPositon(r, poolId, types.MinInitializedTick, types.MaxTick, 1)
}

func (s *KeeperTestSuite) addRandomPositon(r *rand.Rand, poolId uint64, minTick, maxTick int64, tickSpacing int64) {
	tokenDesired0 := sdk.NewCoin(ETH, osmomath.NewInt(r.Int63n(maxAmountDeposited)))
	tokenDesired1 := sdk.NewCoin(USDC, osmomath.NewInt(r.Int63n(maxAmountDeposited)))
	tokensDesired := sdk.NewCoins(tokenDesired0, tokenDesired1)

	accountIndex := r.Intn(len(s.TestAccs))

	s.FundAcc(s.TestAccs[accountIndex], tokensDesired)

	lowerTick := roundTickDownSpacing(r.Int63n(maxTick-minTick+1)+minTick, tickSpacing)
	// lowerTick <= upperTick <= maxTick
	upperTick := roundTickDownSpacing(maxTick-r.Int63n(int64(math.Abs(float64(maxTick-lowerTick)))), tickSpacing)

	fmt.Println("creating position: ", "accountName", "lowerTick", lowerTick, "upperTick", upperTick, "token0Desired", tokenDesired0, "tokenDesired1", tokenDesired1)

	positionData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, s.TestAccs[accountIndex], tokensDesired, osmomath.ZeroInt(), osmomath.ZeroInt(), types.MinInitializedTick, types.MaxTick)
	s.Require().NoError(err)
	fmt.Printf("actually created: %s%s %s%s \n", positionData.Amount0, ETH, positionData.Amount1, USDC)

	s.positionData = append(s.positionData, positionAndLiquidity{
		positionId:   positionData.ID,
		liquidity:    positionData.Liquidity,
		accountIndex: accountIndex,
	})
}

func (s *KeeperTestSuite) removeRandomPosition(r *rand.Rand) {
	if len(s.positionData) == 0 {
		return
	}

	positionIndexToRemove := r.Intn(len(s.positionData))
	positionData := s.positionData[positionIndexToRemove]

	positionIdToRemove := positionData.positionId

	withdrawMultiplier := s.choosePartialOrFullWithdraw(r)

	liqToWithdraw := positionData.liquidity

	liqToWithdrawAfterMultiplier := liqToWithdraw.Mul(withdrawMultiplier)

	// Only apply multiplier if it does not make the liquidity be zero
	if !liqToWithdrawAfterMultiplier.IsZero() {
		liqToWithdraw = liqToWithdrawAfterMultiplier
	}

	fmt.Println("withdrawing position: ", "position id", positionIdToRemove, "amtToWithdraw", liqToWithdraw)

	_, _, err := s.App.ConcentratedLiquidityKeeper.WithdrawPosition(s.Ctx, s.TestAccs[positionData.accountIndex], positionIdToRemove, liqToWithdraw)
	s.Require().NoError(err)

	s.positionData[positionIndexToRemove].liquidity = positionData.liquidity.Sub(liqToWithdraw)

	// if full withdraw, remove position from slice
	if s.positionData[positionIndexToRemove].liquidity.IsZero() {
		s.positionData = append(s.positionData[:positionIndexToRemove], s.positionData[positionIndexToRemove+1:]...)
	}
}

func (s *KeeperTestSuite) transferRandomPosition(r *rand.Rand) {
	if len(s.positionData) == 0 {
		return
	}

	positionIndexToRemove := r.Intn(len(s.positionData))
	positionData := s.positionData[positionIndexToRemove]

	positionIdToTransfer := positionData.positionId

	liqToTransfer := positionData.liquidity

	originalOwner := s.TestAccs[positionData.accountIndex]

	newOwnerIndex := r.Intn(len(s.TestAccs))
	newOwner := s.TestAccs[newOwnerIndex]

	if originalOwner.Equals(newOwner) {
		return
	}

	fmt.Println("transferring position: ", "position id", positionIdToTransfer, "liqToTransfer", liqToTransfer, "from account: ", originalOwner.String(), "to account: ", newOwner.String())

	err := s.App.ConcentratedLiquidityKeeper.TransferPositions(s.Ctx, []uint64{positionIdToTransfer}, originalOwner, newOwner)
	s.Require().NoError(err)

	// remove position from slice
	s.positionData = append(s.positionData[:positionIndexToRemove], s.positionData[positionIndexToRemove+1:]...)

	s.positionData = append(s.positionData, positionAndLiquidity{
		positionId:   positionIdToTransfer,
		liquidity:    liqToTransfer,
		accountIndex: newOwnerIndex,
	})
}

// returns multiplier of the liqudity to withdraw
func (s *KeeperTestSuite) choosePartialOrFullWithdraw(r *rand.Rand) osmomath.Dec {
	multiplier := osmomath.OneDec()
	if r.Intn(2) == 0 {
		// full withdraw
		return multiplier
	}

	// partial withdraw
	multiplier = multiplier.Mul(osmomath.NewDec(r.Int63n(100))).QuoInt64(100)

	return multiplier
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

func randomIntAmount(r *rand.Rand) osmomath.Int {
	return osmomath.NewInt(r.Int63n(maxAmountDeposited))
}
