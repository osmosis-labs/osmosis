package types

const (
	// ModuleName is the name of the market module
	ModuleName = "market"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the market module
	RouterKey = ModuleName

	// QuerierRoute is the query router key for the market module
	QuerierRoute = ModuleName
)

// Keys for market store
// Items are stored with the following key: values
//
// - 0x01: sdk.Dec
var (
	// Keys for store prefixed
	TerraPoolDeltaKey = []byte{0x01} // key for terra pool delta which gap between MintPool from BasePool
)
