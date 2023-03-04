package swapstrategy

import sdk "github.com/cosmos/cosmos-sdk/types"

func ComputeFeeChargePerSwapStepOutGivenIn(hasReachedTarget bool, amountIn, amountSpecifiedRemaining, swapFee sdk.Dec) sdk.Dec {
	return computeFeeChargePerSwapStepOutGivenIn(hasReachedTarget, amountIn, amountSpecifiedRemaining, swapFee)
}
