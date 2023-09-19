package swapstrategy

import "github.com/osmosis-labs/osmosis/osmomath"

func ComputeSpreadRewardChargePerSwapStepOutGivenIn(hasReachedTarget bool, amountIn, amountSpecifiedRemaining, spreadFactor osmomath.Dec) osmomath.Dec {
	return computeSpreadRewardChargePerSwapStepOutGivenIn(hasReachedTarget, amountIn, amountSpecifiedRemaining, spreadFactor)
}

func ComputeSpreadRewardChargeFromAmountIn(amountIn, spreadFactor osmomath.Dec) osmomath.Dec {
	return computeSpreadRewardChargeFromAmountIn(amountIn, spreadFactor)
}
