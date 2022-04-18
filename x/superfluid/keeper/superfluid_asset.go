package keeper

import (
	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"

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
	k.SetOsmoEquivalentMultiplier(ctx, epochNum, asset.Denom, sdk.ZeroDec())
	k.DeleteSuperfluidAsset(ctx, asset.Denom)
}

// Returns amount * (1 - k.RiskFactor(asset))
// Fow now, the risk factor is a global constant.
// It will move towards per pool functions.
func (k Keeper) GetRiskAdjustedOsmoValue(ctx sdk.Context, asset types.SuperfluidAsset, amount sdk.Int) sdk.Int {
	minRiskFactor := k.GetParams(ctx).MinimumRiskFactor
	return amount.Sub(amount.ToDec().Mul(minRiskFactor).RoundInt())
}

// y = x - (x * minRisk)
// y = x (1 - minRisk)
// y / (1 - minRisk) = x

func (k Keeper) UnriskAdjustOsmoValue(ctx sdk.Context, amount sdk.Dec) sdk.Dec {
	minRiskFactor := k.GetParams(ctx).MinimumRiskFactor
	return amount.Quo(sdk.OneDec().Sub(minRiskFactor))
}

func (k Keeper) AddNewSuperfluidAsset(ctx sdk.Context, asset types.SuperfluidAsset) {
	// initialize osmo equivalent multipliers
	epochIdentifier := k.GetEpochIdentifier(ctx)
	currentEpoch := k.ek.GetEpochInfo(ctx, epochIdentifier).CurrentEpoch
	_ = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		k.SetSuperfluidAsset(ctx, asset)
		err := k.UpdateOsmoEquivalentMultipliers(ctx, asset, currentEpoch)
		return err
	})
}
