package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

type ExpectedGlobalRewardValues struct {
	ExpectedAdditiveSpreadRewardTolerance osmomath.Dec
	ExpectedAdditiveIncentivesTolerance   osmomath.Dec
	TotalSpreadRewards                    sdk.Coins
	TotalIncentives                       sdk.Coins
}

// assertGlobalInvariants asserts all available global invariants (i.e. invariants that should hold on all valid states).
// Does not persist any changes to state.
func (s *KeeperTestSuite) assertGlobalInvariants(expectedGlobalRewardValues ExpectedGlobalRewardValues) {
	s.assertTotalRewardsInvariant(expectedGlobalRewardValues)
	s.assertWithdrawAllInvariant()
}

// getAllPositionsAndBalances returns all the positions in state alongside all the pool balances for all pools in state.
//
// Returns:
// * All positions across all pools
// * Total pool assets across all pools
// * Total pool spread rewards across all pools
// * Total pool incentives across all pools
func (s *KeeperTestSuite) getAllPositionsAndPoolBalances(ctx sdk.Context) ([]model.Position, sdk.Coins, sdk.Coins, sdk.Coins) {
	// Get total spread rewards distributed to all pools
	allPools, err := s.Clk.GetPools(ctx)
	totalPoolAssets, totalSpreadRewards, totalIncentives := sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins()

	// Sum up pool balances across all pools
	for _, pool := range allPools {
		clPool, ok := pool.(types.ConcentratedPoolExtension)
		s.Require().True(ok)
		totalPoolAssets = totalPoolAssets.Add(s.App.BankKeeper.GetBalance(ctx, clPool.GetAddress(), clPool.GetToken0()))
		totalPoolAssets = totalPoolAssets.Add(s.App.BankKeeper.GetBalance(ctx, clPool.GetAddress(), clPool.GetToken1()))
		totalSpreadRewards = totalSpreadRewards.Add(s.App.BankKeeper.GetBalance(ctx, clPool.GetSpreadRewardsAddress(), clPool.GetToken0()))
		totalSpreadRewards = totalSpreadRewards.Add(s.App.BankKeeper.GetBalance(ctx, clPool.GetSpreadRewardsAddress(), clPool.GetToken1()))
		totalIncentives = totalIncentives.Add(s.App.BankKeeper.GetAllBalances(ctx, clPool.GetIncentivesAddress())...)
	}

	// Get all positions in state
	allPoolPositions, err := s.Clk.GetAllPositions(ctx)
	s.Require().NoError(err)

	return allPoolPositions, totalPoolAssets, totalSpreadRewards, totalIncentives
}

// assertTotalRewardsInvariant asserts two invariants on the current context:
// 1. Claiming spread rewards and incentives for all positions in state yields the amount stored in pool reward addresses minus rounding errors
// 2. Claiming spread rewards and incentives for all positions in state empties all pool reward addresses except for rounding errors
//
// This function operates on cached context to avoid persisting any changes to state.
func (s *KeeperTestSuite) assertTotalRewardsInvariant(expectedGlobalRewardValues ExpectedGlobalRewardValues) {
	// Get all positions and total pool balances across all CL pools in state
	allPositions, initialTotalPoolLiquidity, expectedTotalSpreadRewards, expectedTotalIncentives := s.getAllPositionsAndPoolBalances(s.Ctx)

	if expectedGlobalRewardValues.TotalSpreadRewards != nil {
		expectedTotalSpreadRewards = expectedGlobalRewardValues.TotalSpreadRewards
	}

	if expectedGlobalRewardValues.TotalIncentives != nil {
		expectedTotalIncentives = expectedGlobalRewardValues.TotalIncentives
	}

	// Switch to cached context to avoid persisting any changes to state
	cachedCtx, _ := s.Ctx.CacheContext()

	// Collect spread rewards for all positions and track output
	totalCollectedSpread, totalCollectedIncentives := sdk.NewCoins(), sdk.NewCoins()
	for _, position := range allPositions {
		owner, err := sdk.AccAddressFromBech32(position.Address)
		s.Require().NoError(err)

		// Log initial position owner balance
		initialBalance := s.App.BankKeeper.GetAllBalances(cachedCtx, owner)

		// Collect spread rewards.
		collectedSpread, err := s.Clk.CollectSpreadRewards(cachedCtx, owner, position.PositionId)
		s.Require().NoError(err)

		// Collect incentives.
		//
		// Since we expect forfeited coins to go to other positions who have not yet claimed, we
		// do not include them in the sum.
		//
		// Balancer full range incentives are also not factored in because they are claimed and sent to
		// gauge immediately upon distribution.
		collectedIncentives, _, _, err := s.Clk.CollectIncentives(cachedCtx, owner, position.PositionId)
		s.Require().NoError(err)

		// Ensure position owner's balance was updated correctly
		finalBalance := s.App.BankKeeper.GetAllBalances(cachedCtx, owner)
		s.Require().Equal(initialBalance.Add(collectedSpread...).Add(collectedIncentives...), finalBalance)

		// Track total amounts
		totalCollectedSpread = totalCollectedSpread.Add(collectedSpread...)
		totalCollectedIncentives = totalCollectedIncentives.Add(collectedIncentives...)
	}

	spreadRewardAdditiveTolerance := osmomath.Dec{}
	if !expectedGlobalRewardValues.ExpectedAdditiveSpreadRewardTolerance.IsNil() {
		spreadRewardAdditiveTolerance = expectedGlobalRewardValues.ExpectedAdditiveSpreadRewardTolerance
	}

	incentivesAdditiveTolerance := osmomath.Dec{}
	if !expectedGlobalRewardValues.ExpectedAdditiveIncentivesTolerance.IsNil() {
		incentivesAdditiveTolerance = expectedGlobalRewardValues.ExpectedAdditiveSpreadRewardTolerance
	}

	// We ensure that any rounding error was in the pool's favor by rounding down.
	// This is to allow for cases where we slightly overround, which would otherwise fail here.
	// TODO: multiplicative tolerance to allow for
	// tightening this check further.
	spreadRewardErrTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: spreadRewardAdditiveTolerance,
		RoundingDir:       osmomath.RoundDown,
	}

	incentivesErrTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: incentivesAdditiveTolerance,
		RoundingDir:       osmomath.RoundDown,
	}

	// Assert total collected spread rewards and incentives equal to expected
	s.Require().True(spreadRewardErrTolerance.EqualCoins(expectedTotalSpreadRewards, totalCollectedSpread), "expected spread rewards vs. collected: %s vs. %s", expectedTotalSpreadRewards, totalCollectedSpread)
	s.Require().True(incentivesErrTolerance.EqualCoins(expectedTotalIncentives, totalCollectedIncentives), "expected incentives vs. collected: %s vs. %s", expectedTotalIncentives, totalCollectedIncentives)

	// Refetch total pool balances across all pools
	remainingPositions, finalTotalPoolLiquidity, remainingTotalSpreadRewards, remainingTotalIncentives := s.getAllPositionsAndPoolBalances(cachedCtx)

	// Ensure pool liquidity remains unchanged
	s.Require().Equal(initialTotalPoolLiquidity, finalTotalPoolLiquidity)

	// Ensure total remaining spread rewards and incentives are exactly equal to loss due to rounding
	if expectedGlobalRewardValues.TotalSpreadRewards == nil {
		roundingLossSpread := expectedTotalSpreadRewards.Sub(totalCollectedSpread...)
		s.Require().Equal(roundingLossSpread, remainingTotalSpreadRewards)
	}

	if expectedGlobalRewardValues.TotalIncentives == nil {
		roundingLossIncentives := expectedTotalIncentives.Sub(totalCollectedIncentives...)
		s.Require().Equal(roundingLossIncentives, remainingTotalIncentives)
	}

	// Ensure no positions were deleted
	s.Require().Equal(len(allPositions), len(remainingPositions))
}

// assertWithdrawAllInvariant withdraws all positions from all pools in state and asserts that all pool liquidity was removed from pool balances.
func (s *KeeperTestSuite) assertWithdrawAllInvariant() {
	// Get all positions and pool balances across all CL pools in state
	allPositions, expectedTotalWithdrawn, _, _ := s.getAllPositionsAndPoolBalances(s.Ctx)

	// Switch to cached context to avoid persisting any changes to state
	cachedCtx, _ := s.Ctx.CacheContext()

	// Withdraw all assets for all positions and track output
	totalWithdrawn := sdk.NewCoins()
	for _, position := range allPositions {
		owner, err := sdk.AccAddressFromBech32(position.Address)
		s.Require().NoError(err)

		// Withdraw all assets from position
		amt0Withdrawn, amt1Withdrawn, err := s.Clk.WithdrawPosition(cachedCtx, owner, position.PositionId, position.Liquidity)
		s.Require().NoError(err)

		// Convert withdrawn assets to coins
		positionPool, err := s.Clk.GetPoolById(cachedCtx, position.PoolId)
		s.Require().NoError(err)
		withdrawn := sdk.NewCoins(
			sdk.NewCoin(positionPool.GetToken0(), amt0Withdrawn),
			sdk.NewCoin(positionPool.GetToken1(), amt1Withdrawn),
		)

		// Track total withdrawn assets
		totalWithdrawn = totalWithdrawn.Add(withdrawn...)
	}

	// For global invariant checks, we simply ensure that any rounding error was in the pool's favor.
	// This is to allow for cases where we slightly overround, which would otherwise fail here.
	// TODO: create ErrTolerance type that allows for additive OR multiplicative tolerance to allow for
	// tightening this check further.
	errTolerance := osmomath.ErrTolerance{
		RoundingDir: osmomath.RoundDown,
	}

	// Assert total withdrawn assets equal to expected
	s.Require().True(errTolerance.EqualCoins(expectedTotalWithdrawn, totalWithdrawn), "expected withdrawn vs. actual: %s vs. %s", expectedTotalWithdrawn, totalWithdrawn)

	// Refetch total pool balances across all pools
	remainingPositions, finalTotalPoolAssets, remainingTotalSpreadRewards, remainingTotalIncentives := s.getAllPositionsAndPoolBalances(cachedCtx)

	// Ensure no more positions exist in state
	s.Require().Equal(0, len(remainingPositions))

	// Ensure pool liquidity only has rounding error left in it
	roundingLossAssets := expectedTotalWithdrawn.Sub(totalWithdrawn...)
	s.Require().Equal(roundingLossAssets, finalTotalPoolAssets)

	// Ensure spread rewards and incentives are all claimed except for rounding error
	s.Require().True(errTolerance.EqualCoins(remainingTotalSpreadRewards, sdk.NewCoins()))
	s.Require().True(errTolerance.EqualCoins(remainingTotalIncentives, sdk.NewCoins()))
}
