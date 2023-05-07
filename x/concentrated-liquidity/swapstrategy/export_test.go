package swapstrategy

import sdk "github.com/cosmos/cosmos-sdk/types"

func ComputeFeeChargePerSwapStepOutGivenIn(hasReachedTarget bool, amountIn, amountSpecifiedRemaining, swapFee sdk.Dec) sdk.Dec {
	return computeFeeChargePerSwapStepOutGivenIn(hasReachedTarget, amountIn, amountSpecifiedRemaining, swapFee)
}

func ComputeFeeChargeFromAmountIn(amountIn, swapFee sdk.Dec) sdk.Dec {
	return computeFeeChargeFromAmountIn(amountIn, swapFee)
}
