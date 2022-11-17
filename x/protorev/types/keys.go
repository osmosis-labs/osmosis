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
	prefixSearcherRoutes = iota + 1
	prefixOsmoPools
	prefixAtomPools
	prefixProtoRevStatistics
)

var (
	// KeyPrefixSearcherRoutes is the prefix for the SearcherRoutes store
	KeyPrefixSearcherRoutes = []byte{prefixSearcherRoutes}

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

// Returns the key need to fetch the searcher routes for a given pool id
func GetKeyPrefixRouteForTokenPair(tokenA, tokenB string) []byte {
	// sort the tokens after converting to upper case
	upperA := strings.ToUpper(tokenA)
	upperB := strings.ToUpper(tokenB)

	if upperA < upperB {
		return append(KeyPrefixSearcherRoutes, []byte(upperA+upperB)...)
	}
	return append(KeyPrefixSearcherRoutes, []byte(upperB+upperA)...)
}
