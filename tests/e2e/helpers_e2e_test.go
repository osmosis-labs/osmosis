package e2e

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/tests/e2e/configurer/chain"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
)

var defaultFeePerTx = sdk.NewInt(1000)

// calculateSpreadRewardGrowthGlobal calculates spread reward growth global per unit of virtual liquidity based on swap parameters:
// amountIn - amount being swapped
// spreadFactor - pool's spread factor
// poolLiquidity - current pool liquidity
func calculateSpreadRewardGrowthGlobal(amountIn, spreadFactor, poolLiquidity sdk.Dec) sdk.Dec {
	// First we get total spread reward charge for the swap (ΔY * spreadFactor)
	spreadRewardChargeTotal := amountIn.Mul(spreadFactor)

	// Calculating spread reward growth global (dividing by pool liquidity to find spread reward growth per unit of virtual liquidity)
	spreadRewardGrowthGlobal := spreadRewardChargeTotal.Quo(poolLiquidity)
	return spreadRewardGrowthGlobal
}

// calculateSpreadRewardGrowthInside calculates spread reward growth inside range per unit of virtual liquidity
// spreadRewardGrowthGlobal - global spread reward growth per unit of virtual liquidity
// spreadRewardGrowthBelow - spread reward growth below lower tick
// spreadRewardGrowthAbove - spread reward growth above upper tick
// Formula: spreadRewardGrowthGlobal - spreadRewardGrowthBelowLowerTick - spreadRewardGrowthAboveUpperTick
func calculateSpreadRewardGrowthInside(spreadRewardGrowthGlobal, spreadRewardGrowthBelow, spreadRewardGrowthAbove sdk.Dec) sdk.Dec {
	return spreadRewardGrowthGlobal.Sub(spreadRewardGrowthBelow).Sub(spreadRewardGrowthAbove)
}

// Assert balances that are not affected by swap:
// * same amount of `stake` in balancesBefore and balancesAfter
// * amount of `e2e-default-feetoken` dropped by 1000 (default amount for fee per tx)
// * depending on `assertUosmoBalanceIsConstant` and `assertUionBalanceIsConstant` parameters, check that those balances have also not been changed
func (s *IntegrationTestSuite) assertBalancesInvariants(balancesBefore, balancesAfter sdk.Coins, assertUosmoBalanceIsConstant, assertUionBalanceIsConstant bool) {
	s.Require().True(balancesAfter.AmountOf("stake").Equal(balancesBefore.AmountOf("stake")))
	s.Require().True(balancesAfter.AmountOf("e2e-default-feetoken").Equal(balancesBefore.AmountOf("e2e-default-feetoken").Sub(defaultFeePerTx)))
	if assertUionBalanceIsConstant {
		s.Require().True(balancesAfter.AmountOf("uion").Equal(balancesBefore.AmountOf("uion")))
	}
	if assertUosmoBalanceIsConstant {
		s.Require().True(balancesAfter.AmountOf("uosmo").Equal(balancesBefore.AmountOf("uosmo")))
	}
}

// Get balances for address
func (s *IntegrationTestSuite) addrBalance(node *chain.NodeConfig, address string) sdk.Coins {
	addrBalances, err := node.QueryBalances(address)
	s.Require().NoError(err)
	return addrBalances
}

// Helper function for calculating uncollected spread rewards since the time that spreadRewardGrowthInsideLast corresponds to
// positionLiquidity - current position liquidity
// spreadRewardGrowthBelow - spread reward growth below lower tick
// spreadRewardGrowthAbove - spread reward growth above upper tick
// spreadRewardGrowthInsideLast - amount of spread reward growth inside range at the time from which we want to calculate the amount of uncollected spread rewards
// spreadRewardGrowthGlobal - variable for tracking global spread reward growth
func calculateUncollectedSpreadRewards(positionLiquidity, spreadRewardGrowthBelow, spreadRewardGrowthAbove, spreadRewardGrowthInsideLast sdk.Dec, spreadRewardGrowthGlobal sdk.Dec) sdk.Dec {
	// Calculating spread reward growth inside range [-1200; 400]
	spreadRewardGrowthInside := calculateSpreadRewardGrowthInside(spreadRewardGrowthGlobal, spreadRewardGrowthBelow, spreadRewardGrowthAbove)

	// Calculating uncollected spread rewards
	// Formula for finding uncollected spread rewards in time range [t1; t2]:
	// F_u = position_liquidity * (spread_rewards_growth_inside_t2 - spread_rewards_growth_inside_t1).
	spreadRewardsUncollected := positionLiquidity.Mul(spreadRewardGrowthInside.Sub(spreadRewardGrowthInsideLast))

	return spreadRewardsUncollected
}

// Get current (updated) pool
func (s *IntegrationTestSuite) updatedConcentratedPool(node *chain.NodeConfig, poolId uint64) types.ConcentratedPoolExtension {
	concentratedPool, err := node.QueryConcentratedPool(poolId)
	s.Require().NoError(err)
	return concentratedPool
}

func (s *IntegrationTestSuite) updatedCFMMPool(node *chain.NodeConfig, poolId uint64) gammtypes.CFMMPoolI {
	cfmmPool, err := node.QueryCFMMPool(poolId)
	s.Require().NoError(err)
	return cfmmPool
}

// Assert returned positions:
func (s *IntegrationTestSuite) validateCLPosition(position model.Position, poolId uint64, lowerTick, upperTick int64) {
	s.Require().Equal(position.PoolId, poolId)
	s.Require().Equal(position.LowerTick, lowerTick)
	s.Require().Equal(position.UpperTick, upperTick)
}
