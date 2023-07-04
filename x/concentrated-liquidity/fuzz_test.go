package concentrated_liquidity_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

func (s *KeeperTestSuite) TestFuzz() {
	spreadFactor := sdk.NewDecWithPrec(1, 3)
	pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, spreadFactor)
	defaultCoins := sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(1_000_000)), sdk.NewCoin(USDC, sdk.NewInt(1_000_000_000)))
	s.CreateFullRangePosition(pool, defaultCoins)
	s.FuzzTest(pool.GetId(), 100, 20)
}

// pre-condition: poolId exists, and has at least one position
func (s *KeeperTestSuite) FuzzTest(poolId uint64, numSwaps int, numPositions int) {
	r := rand.New(rand.NewSource(rand.Int63()))
	s.fuzzTestWithSeed(r, poolId, numSwaps, numPositions)
}

type fuzzState struct {
	r      *rand.Rand
	poolId int
}

func (s *KeeperTestSuite) fuzzTestWithSeed(r *rand.Rand, poolId uint64, numSwaps int, numPositions int) {
	completedSwaps := 0
	completedPositions := 0
	targetActions := numSwaps + numPositions
	for i := 0; i < targetActions; i++ {
		doSwap := selectAction(r, numSwaps, numPositions, completedSwaps, completedPositions)
		if doSwap {
			// s.swap(r)
			s.randomSwap(r, poolId)
			completedSwaps++
		} else {
			// s.addOrRemoveLiquidity(r)
			completedPositions++
		}
	}
}

func (s *KeeperTestSuite) randomSwap(r *rand.Rand, poolId uint64) {
	// High level decision, decide which swap strategy to do.
	// 1. Swap a random amount
	// 2. Swap near next tick boundary
	// 3. Swap to later tick boundary (TODO)
	swapStrategy := r.Intn(2)
	if swapStrategy == 0 {
		// s.swapRandomAmount(r)
	} else {
		// s.swapNearNextTickBoundary(r)
	}
}

func (s *KeeperTestSuite) swapRandomAmount(r *rand.Rand, poolId uint64) {}

func (s *KeeperTestSuite) swapNearNextTickBoundary(r *rand.Rand, poolId uint64) {
	pool, _ := s.clk.GetPool(s.Ctx, poolId)
	clPool := pool.(types.ConcentratedPoolExtension)

	zfo := s.chooseSwapDirection(r, clPool)
	targetTick := clPool.GetCurrentTick()
	if zfo {
		targetTick -= 1
	} else {
		targetTick += 1
	}
	s.swapNearTickBoundary(r, clPool, targetTick, zfo)
}

func (s *KeeperTestSuite) swapNearTickBoundary(r *rand.Rand, pool types.ConcentratedPoolExtension, targetTick int64, zfo bool) {
	swapInDenom, swapOutDenom := zfoToDenoms(zfo, pool)
	// TODO: Confirm accuracy of this method.
	amountInRequired, _, _ := s.computeSwapAmounts(pool.GetId(), pool.GetCurrentSqrtPrice(), targetTick, zfo, false)

	_, _, _ = swapInDenom, swapOutDenom, amountInRequired
	// TODO: What is all this?
	// poolSpotPrice := pool.GetCurrentSqrtPrice().Power(osmomath.NewBigDec(2))
	// minSwapOutAmount := poolSpotPrice.Mul(osmomath.SmallestDec()).SDKDec().TruncateInt()
	// poolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
	// if poolBalances.AmountOf(swapOutDenom).LTE(minSwapOutAmount) {
	// 	fmt.Println("skipped")
	// 	return sdk.Coin{}, sdk.Coin{}
	// }

	// fmt.Println("dec amt in required to get to tick boundary: ", amountInRequired)
	// swapInFunded := sdk.NewCoin(swapInDenom, amountInRequired.TruncateInt())
	// s.FundAcc(swapAddress, sdk.NewCoins(swapInFunded))

	// // Execute swap
	// fmt.Println("begin keeper swap")
	// swappedIn, swappedOut, _, err := s.clk.SwapOutAmtGivenIn(s.Ctx, swapAddress, pool, swapInFunded, swapOutDenom, pool.GetSpreadFactor(s.Ctx), sdk.ZeroDec())
	// if errors.As(err, &types.InvalidAmountCalculatedError{}) {
	// 	// If the swap we're about to execute will not generate enough output, we skip the swap.
	// 	// it would error for a real user though. This is good though, since that user would just be burning funds.
	// 	if err.(types.InvalidAmountCalculatedError).Amount.IsZero() {
	// 		fmt.Println("Hit error for 0 out, swap failed")
	// 		fmt.Println("TODO: Revert swap attempt")
	// 		return sdk.Coin{}, sdk.Coin{}
	// 	} else {
	// 		s.Require().NoError(err)
	// 	}
	// } else {
	// 	s.Require().NoError(err)
	// }

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

func selectAction(r *rand.Rand, numSwaps, numPositions, completedSwaps, completedPositions int) bool {
	if completedSwaps == numSwaps {
		return false
	}
	if completedPositions == numPositions {
		return true
	}
	return r.Intn(2) == 0
}
