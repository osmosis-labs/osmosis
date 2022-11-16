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
	return append(KeyPrefixOsmoPools, []byte(denom)...)
}

// Returns the key needed to fetch the atom pool for a given denom
func GetKeyPrefixAtomPool(denom string) []byte {
	return append(KeyPrefixAtomPools, []byte(denom)...)
}

// Returns the key need to fetch the route for a given pair of denoms
func GetKeyPrefixRouteForPair(denom1, denom2 string) []byte {
	first := strings.ToUpper(denom1)
	second := strings.ToUpper(denom2)

	if denom1 < denom2 {
		first = denom1
		second = denom2
	} else {
		first = denom2
		second = denom1
	}

	return append(KeyPrefixRoutes, []byte(first+second)...)
}
