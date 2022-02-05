package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewSuperfluidAsset returns a new instance of SuperfluidAsset
func NewSuperfluidAsset(assetType SuperfluidAssetType, denom string) SuperfluidAsset {
	return SuperfluidAsset{
		AssetType: assetType,
		Denom:     denom,
	}
}

func (a SuperfluidIntermediaryAccount) GetAddress() sdk.AccAddress {
	return authtypes.NewModuleAddress(a.Denom + a.ValAddr)
}
