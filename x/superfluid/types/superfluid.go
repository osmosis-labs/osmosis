package types

// NewSuperfluidAsset returns a new instance of SuperfluidAsset
func NewSuperfluidAsset(assetType SuperfluidAssetType, denom string) SuperfluidAsset {
	return SuperfluidAsset{
		AssetType: assetType,
		Denom:     denom,
	}
}
