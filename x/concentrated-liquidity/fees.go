package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Add adds amount to the first or second token depending on the boolean first
func (f *Fee) Add(first bool, amount sdk.Dec) sdk.Dec {
	if first {
		return f.Token0.Add(amount)
	} else {
		return f.Token1.Add(amount)
	}
}

// Sub returns a new fee with the amount of token0 and token1 subtracted respectively
func (f *Fee) Sub(fee *Fee) *Fee {
	newFee := &Fee{}
	newFee.Token0 = f.Token0.Sub(fee.Token0)
	newFee.Token1 = f.Token1.Sub(fee.Token1)
	return newFee
}

// UpdateFeesForTick updates all initialized ticks below the current tick with the fees accrued in the current tick
// ToDo: Does it matter that we skip uninitialized ticks?
func (k Keeper) UpdateFeesForTick(ctx sdk.Context, poolId uint64, tickIndex int64, fee sdk.Dec, firstToken bool) error {
	tick, initialized := k.NextInitializedTick(ctx, poolId, tickIndex, true)
	for initialized {
		tickInfo, err := k.GetTickInfo(ctx, poolId, tick)
		if err != nil {
			return err
		}

		tickInfo.FeesAccruedAboveTick.Add(firstToken, fee)
		k.setTickInfo(ctx, poolId, tickIndex, tickInfo)
	}
	return nil
}

func (k Keeper) ClaimFees(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64) (*Fee, error) {
	pool := k.getPoolbyId(ctx, poolId)
	position, err := k.GetPosition(ctx, poolId, owner, lowerTick, upperTick)
	if err != nil {
		return nil, err
	}
	upperTickInfo, err := k.GetTickInfo(ctx, poolId, upperTick)
	if err != nil {
		return nil, err
	}
	lowerTickInfo, err := k.GetTickInfo(ctx, poolId, lowerTick)
	if err != nil {
		return nil, err
	}

	totalFees := upperTickInfo.FeesAccruedAboveTick.Sub(lowerTickInfo.FeesAccruedAboveTick).Sub(position.FeesAtCreation)

	pool.GlobalFees.Sub(totalFees)

	feeShares := totalFees // ToDo: .Quo(owner.shares)

	// ToDo: Check for tick overflow? is upperTick+1 correct?
	err = k.UpdateFeesForTick(ctx, poolId, upperTick+1, feeShares.Token0.Neg(), true)
	if err != nil {
		return nil, err
	}
	err = k.UpdateFeesForTick(ctx, poolId, upperTick+1, feeShares.Token1.Neg(), false)
	if err != nil {
		return nil, err
	}

	return feeShares, nil
}

func ComputeStepFee(totalFee sdk.Dec, ratio sdk.Dec) sdk.Dec {
	return totalFee.Mul(ratio)
}
