package swapstrategy

import "github.com/osmosis-labs/osmosis/osmomath"

func ComputeSpreadRewardChargePerSwapStepOutGivenIn(hasReachedTarget bool, amountIn, amountSpecifiedRemaining, spreadFactor osmomath.Dec) osmomath.Dec {
	oneMinusSpreadFactorGetter := func() osmomath.Dec {
		return osmomath.OneDec().Sub(spreadFactor)
	}
	return computeSpreadRewardChargePerSwapStepOutGivenIn(hasReachedTarget, amountIn, amountSpecifiedRemaining, spreadFactor, oneMinusSpreadFactorGetter)
}

func ComputeSpreadRewardChargeFromAmountIn(amountIn, spreadFactor osmomath.Dec) osmomath.Dec {
	return computeSpreadRewardChargeFromAmountIn(amountIn, spreadFactor, oneDec.Sub(spreadFactor))
}
