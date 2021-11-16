package types

var (
	// ModuleName defines the module name
	ModuleName = "superfluid"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// KeyPrefixSuperfluidAsset defines prefix key for superfluid asset
	KeyPrefixSuperfluidAsset = []byte{0x01}

	// KeyPrefixSuperfluidAssetInfo defines prefix key for superfluid asset info
	KeyPrefixSuperfluidAssetInfo = []byte{0x02}
)
