package swapstrategy

import "github.com/osmosis-labs/osmosis/osmomath"

func ComputeSpreadRewardChargePerSwapStepOutGivenIn(hasReachedTarget bool, amountIn, amountSpecifiedRemaining, spreadFactor osmomath.Dec) osmomath.Dec {
	spreadFactorOverOneMinusSpreadFactorGetter := func() osmomath.Dec {
		return spreadFactor.QuoRoundUp(oneDec.Sub(spreadFactor))
	}
	return computeSpreadRewardChargePerSwapStepOutGivenIn(hasReachedTarget, amountIn, amountSpecifiedRemaining, spreadFactor, spreadFactorOverOneMinusSpreadFactorGetter)
}

func ComputeSpreadRewardChargeFromAmountIn(amountIn, spreadFactor osmomath.Dec) osmomath.Dec {
	return computeSpreadRewardChargeFromAmountIn(amountIn, spreadFactor.QuoRoundUp(oneDec.Sub(spreadFactor)))
}
