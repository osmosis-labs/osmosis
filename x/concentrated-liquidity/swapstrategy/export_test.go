package swapstrategy

import sdk "github.com/cosmos/cosmos-sdk/types"

type SwapStrategy = swapStrategy

func ComputeSpreadRewardChargePerSwapStepOutGivenIn(hasReachedTarget bool, amountIn, amountSpecifiedRemaining, spreadFactor sdk.Dec) sdk.Dec {
	return computeSpreadRewardChargePerSwapStepOutGivenIn(hasReachedTarget, amountIn, amountSpecifiedRemaining, spreadFactor)
}

func ComputeSpreadRewardChargeFromAmountIn(amountIn, spreadFactor sdk.Dec) sdk.Dec {
	return computeSpreadRewardChargeFromAmountIn(amountIn, spreadFactor)
}
