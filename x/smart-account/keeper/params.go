package keeper

import (
	"bytes"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/smart-account/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetIsSmartAccountActive returns the value of the isSmartAccountActive parameter.
// If the value has not been set, it will return false.
// If there is an error unmarshalling the value, it will return false.
func (k *Keeper) GetIsSmartAccountActive(ctx sdk.Context) bool {
	isSmartAccountActiveBz := k.paramSpace.GetRaw(ctx, types.KeyIsSmartAccountActive)
	if !bytes.Equal(isSmartAccountActiveBz, k.isSmartAccountActiveBz) {
		var isSmartAccountActiveValue bool
		err := json.Unmarshal(isSmartAccountActiveBz, &isSmartAccountActiveValue)
		if err != nil {
			k.Logger(ctx).Error("failed to unmarshal isSmartAccountActive", "error", err)
			isSmartAccountActiveValue = false
		}
		k.isSmartAccountActiveVal = isSmartAccountActiveValue
		k.isSmartAccountActiveBz = isSmartAccountActiveBz
	}
	return k.isSmartAccountActiveVal
}
