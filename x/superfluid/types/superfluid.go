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

func NewSuperfluidIntermediaryAccount(denom string, valAddr string, gaugeId uint64) SuperfluidIntermediaryAccount {
	return SuperfluidIntermediaryAccount{
		Denom:   denom,
		ValAddr: valAddr,
		GaugeId: gaugeId,
	}
}

func (a SuperfluidIntermediaryAccount) GetAccAddress() sdk.AccAddress {
	return authtypes.NewModuleAddress(a.Denom + a.ValAddr)
}
