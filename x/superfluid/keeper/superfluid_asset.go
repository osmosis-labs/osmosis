package keeper

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v25/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginUnwindSuperfluidAsset starts the deletion process for a superfluid asset.
// This current method is a stub, but is called when:
// * Governance removes a superfluid asset
// * A severe error in gamm occurs
//
// It should eventually begin unwinding all of the synthetic lockups for that asset
// and queue them for deletion.
// See https://github.com/osmosis-labs/osmosis/issues/864
func (k Keeper) BeginUnwindSuperfluidAsset(ctx sdk.Context, epochNum int64, asset types.SuperfluidAsset) {
	// Right now set the TWAP to 0, and delete the asset.
	k.SetOsmoEquivalentMultiplier(ctx, epochNum, asset.Denom, osmomath.ZeroDec())
	k.DeleteSuperfluidAsset(ctx, asset.Denom)
}

// Returns amount * (1 - k.RiskFactor(asset))
func (k Keeper) GetRiskAdjustedOsmoValue(ctx sdk.Context, amount osmomath.Int, denom string) osmomath.Int {
	riskFactor := k.CalculateRiskFactor(ctx, denom)
	return amount.Sub(amount.ToLegacyDec().Mul(riskFactor).RoundInt())
}

// CalculateRiskFactor Will try to fetch the specific risk factor for the denom, and if it
// doesn't exist, will return the minimum risk factor.
func (k Keeper) CalculateRiskFactor(ctx sdk.Context, denom string) osmomath.Dec {
	if riskFactor, found := k.GetDenomRiskFactor(ctx, denom); found {
		return riskFactor
	}
	return k.GetParams(ctx).MinimumRiskFactor
}

// y = x - (x * riskFactor)
// y = x (1 - riskFactor)
// y / (1 - riskFactor) = x

func (k Keeper) UnriskAdjustOsmoValue(ctx sdk.Context, amount osmomath.Dec, denom string) osmomath.Dec {
	riskFactor := k.CalculateRiskFactor(ctx, denom)
	return amount.Quo(osmomath.OneDec().Sub(riskFactor))
}

func (k Keeper) AddNewSuperfluidAsset(ctx sdk.Context, asset types.SuperfluidAsset) error {
	// initialize osmo equivalent multipliers
	epochIdentifier := k.GetEpochIdentifier(ctx)
	currentEpoch := k.ek.GetEpochInfo(ctx, epochIdentifier).CurrentEpoch
	return osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		k.SetSuperfluidAsset(ctx, asset)
		err := k.UpdateOsmoEquivalentMultipliers(ctx, asset, currentEpoch)
		return err
	})
}
