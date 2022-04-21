package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewSuperfluidAsset returns a new instance of SuperfluidAsset.
func NewSuperfluidAsset(assetType SuperfluidAssetType, denom string) SuperfluidAsset {
	return SuperfluidAsset{
		AssetType: assetType,
		Denom:     denom,
	}
}

func NewSuperfluidIntermediaryAccount(denom string, valAddr string, gaugeId uint64) SuperfluidIntermediaryAccount {
	return SuperfluidIntermediaryAccount{
		Denom:   denom,
		ValAddr: valAddr,
		GaugeId: gaugeId,
	}
}

func (a SuperfluidIntermediaryAccount) Empty() bool {
	// if intermediary account isn't set in state, we get the default intermediary account.
	// if it set, then the denom is non-blank
	return a.Denom == ""
}

func (a SuperfluidIntermediaryAccount) GetAccAddress() sdk.AccAddress {
	return GetSuperfluidIntermediaryAccountAddr(a.Denom, a.ValAddr)
}

func GetSuperfluidIntermediaryAccountAddr(denom, valAddr string) sdk.AccAddress {
	// TODO: Make this better namespaced.
	// if ValAddr's one day switch to potentially be 32 bytes, a malleability attack could be crafted.
	// We are launching with the address as is, so this will have to be done as a migration in the future.
	return authtypes.NewModuleAddress(denom + valAddr)
}
