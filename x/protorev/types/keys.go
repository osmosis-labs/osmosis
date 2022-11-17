package types

import "strings"

const (
	// ModuleName defines the module name
	ModuleName = "protorev"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

const (
	prefixRoute = iota + 1
	prefixOsmoPools
	prefixAtomPools
	prefixProtoRevStatistics
)

var (
	// KeyPrefixRoute is the prefix for the route store
	KeyPrefixRoutes = []byte{prefixRoute}

	// KeyPrefixOsmoPools is the prefix for the osmo pool store
	KeyPrefixOsmoPools = []byte{prefixOsmoPools}

	// KeyPrefixAtomPools is the prefix for the atom pool store
	KeyPrefixAtomPools = []byte{prefixAtomPools}

	// KeyPrefixProtoRevStatistics is the prefix for the proto rev statistics store
	KeyPrefixProtoRevStatistics = []byte{prefixProtoRevStatistics}
)

// Returns the key needed to fetch the osmo pool for a given denom
func GetKeyPrefixOsmoPool(denom string) []byte {
	upper := strings.ToUpper(denom)
	return append(KeyPrefixOsmoPools, []byte(upper)...)
}

// Returns the key needed to fetch the atom pool for a given denom
func GetKeyPrefixAtomPool(denom string) []byte {
	upper := strings.ToUpper(denom)
	return append(KeyPrefixAtomPools, []byte(upper)...)
}

// Returns the key need to fetch the route for a given pair of denoms
func GetKeyPrefixRouteForPoolID(poolID uint64) []byte {
	return append(KeyPrefixRoutes, UInt64ToBytes(poolID)...)
}
