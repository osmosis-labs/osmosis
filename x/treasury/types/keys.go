package types

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "treasury"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the message route for treasury
	RouterKey = ModuleName

	// QuerierRoute is the querier route for treasury
	QuerierRoute = ModuleName
)

// Keys for treasury store
// Items are stored with the following key: values
//
// - 0x01: osmomath.Dec
var (
	// TaxRateKey for store prefixes
	TaxRateKey = []byte{0x01} // a key for a tax-rate
)
