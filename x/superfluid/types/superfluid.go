package types

// NewSuperfluidAsset returns a new instance of SuperfluidAsset
func NewSuperfluidAsset(denom string) SuperfluidAsset {
	return SuperfluidAsset{
		AssetType: SuperfluidAssetTypeDefault,
		Denom:     denom,
	}
}
