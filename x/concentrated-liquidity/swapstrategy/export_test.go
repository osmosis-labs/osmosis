package swapstrategy

import sdk "github.com/cosmos/cosmos-sdk/types"

func ComputespreadRewardChargePerSwapStepOutGivenIn(hasReachedTarget bool, amountIn, amountSpecifiedRemaining, spreadFactor sdk.Dec) sdk.Dec {
	return computespreadRewardChargePerSwapStepOutGivenIn(hasReachedTarget, amountIn, amountSpecifiedRemaining, spreadFactor)
}

func ComputespreadRewardChargeFromAmountIn(amountIn, spreadFactor sdk.Dec) sdk.Dec {
	return computespreadRewardChargeFromAmountIn(amountIn, spreadFactor)
}
