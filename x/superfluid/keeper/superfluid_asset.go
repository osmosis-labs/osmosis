package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
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
	k.SetEpochOsmoEquivalentTWAP(ctx, epochNum, asset.Denom, sdk.ZeroDec())
	k.DeleteSuperfluidAsset(ctx, asset.Denom)
}

func (k Keeper) GetRiskAdjustedOsmoValue(ctx sdk.Context, asset types.SuperfluidAsset, amount sdk.Int) sdk.Int {
	// TODO: we need to figure out how to do this later.
	minRiskFactor := k.GetParams(ctx).MinimumRiskFactor
	return amount.Sub(amount.ToDec().Mul(minRiskFactor).RoundInt())
}
