package e2e

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/tests/e2e/configurer/chain"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var defaultFeePerTx = sdk.NewInt(1000)

// calculateFeeGrowthGlobal calculates fee growth global per unit of virtual liquidity based on swap parameters:
// amountIn - amount being swapped
// swapFee - pool's swap fee
// poolLiquidity - current pool liquidity
func calculateFeeGrowthGlobal(amountIn, swapFee, poolLiquidity sdk.Dec) sdk.Dec {
	// First we get total fee charge for the swap (ΔY * swapFee)
	feeChargeTotal := amountIn.Mul(swapFee)

	// Calculating fee growth global (dividing by pool liquidity to find fee growth per unit of virtual liquidity)
	feeGrowthGlobal := feeChargeTotal.Quo(poolLiquidity)
	return feeGrowthGlobal
}

// calculateFeeGrowthInside calculates fee growth inside range per unit of virtual liquidity
// feeGrowthGlobal - global fee growth per unit of virtual liquidity
// feeGrowthBelow - fee growth below lower tick
// feeGrowthAbove - fee growth above upper tick
// Formula: feeGrowthGlobal - feeGrowthBelowLowerTick - feeGrowthAboveUpperTick
func calculateFeeGrowthInside(feeGrowthGlobal, feeGrowthBelow, feeGrowthAbove sdk.Dec) sdk.Dec {
	return feeGrowthGlobal.Sub(feeGrowthBelow).Sub(feeGrowthAbove)
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

// Helper function for calculating uncollected fees since the time that feeGrowthInsideLast corresponds to
// positionLiquidity - current position liquidity
// feeGrowthBelow - fee growth below lower tick
// feeGrowthAbove - fee growth above upper tick
// feeGrowthInsideLast - amount of fee growth inside range at the time from which we want to calculate the amount of uncollected fees
// feeGrowthGlobal - variable for tracking global fee growth
func calculateUncollectedFees(positionLiquidity, feeGrowthBelow, feeGrowthAbove, feeGrowthInsideLast sdk.Dec, feeGrowthGlobal sdk.Dec) sdk.Dec {
	// Calculating fee growth inside range [-1200; 400]
	feeGrowthInside := calculateFeeGrowthInside(feeGrowthGlobal, feeGrowthBelow, feeGrowthAbove)

	// Calculating uncollected fees
	// Formula for finding uncollected fees in time range [t1; t2]:
	// F_u = position_liquidity * (fee_growth_inside_t2 - fee_growth_inside_t1).
	feesUncollected := positionLiquidity.Mul(feeGrowthInside.Sub(feeGrowthInsideLast))

	return feesUncollected
}

// Get current (updated) pool
func (s *IntegrationTestSuite) updatedPool(node *chain.NodeConfig, poolId uint64) types.ConcentratedPoolExtension {
	concentratedPool, err := node.QueryConcentratedPool(poolId)
	s.Require().NoError(err)
	return concentratedPool
}

// Assert returned positions:
func (s *IntegrationTestSuite) validateCLPosition(position model.Position, poolId uint64, lowerTick, upperTick int64) {
	s.Require().Equal(position.PoolId, poolId)
	s.Require().Equal(position.LowerTick, lowerTick)
	s.Require().Equal(position.UpperTick, upperTick)
}
