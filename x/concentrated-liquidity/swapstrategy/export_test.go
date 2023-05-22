package swapstrategy

import sdk "github.com/cosmos/cosmos-sdk/types"

func ComputeFeeChargePerSwapStepOutGivenIn(hasReachedTarget bool, amountIn, amountSpecifiedRemaining, spreadFactor sdk.Dec) sdk.Dec {
	return computeFeeChargePerSwapStepOutGivenIn(hasReachedTarget, amountIn, amountSpecifiedRemaining, spreadFactor)
}

func ComputeFeeChargeFromAmountIn(amountIn, spreadFactor sdk.Dec) sdk.Dec {
	return computeFeeChargeFromAmountIn(amountIn, spreadFactor)
}
