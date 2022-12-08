package types

const (
	// ModuleName defines the module name
	ModuleName = "protorev"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

const (
	prefixTokenPairRoutes = iota + 1
	prefixOsmoPools
	prefixAtomPools
)

var (
	// -------------- Keys for trading stores -------------- //
	// KeyPrefixTokenPairRoutes is the prefix for the TokenPairArbRoutes store
	KeyPrefixTokenPairRoutes = []byte{prefixTokenPairRoutes}

	// KeyPrefixOsmoPools is the prefix for the osmo pool store
	KeyPrefixOsmoPools = []byte{prefixOsmoPools}

	// KeyPrefixAtomPools is the prefix for the atom pool store
	KeyPrefixAtomPools = []byte{prefixAtomPools}
)

// Returns the key needed to fetch the osmo pool for a given denom
func GetKeyPrefixOsmoPool(denom string) []byte {
	return append(KeyPrefixOsmoPools, []byte(denom)...)
}

// Returns the key needed to fetch the atom pool for a given denom
func GetKeyPrefixAtomPool(denom string) []byte {
	return append(KeyPrefixAtomPools, []byte(denom)...)
}

// Returns the key needed to fetch the tokenPair routes for a given pair of tokens
func GetKeyPrefixRouteForTokenPair(tokenA, tokenB string) []byte {
	return append(KeyPrefixTokenPairRoutes, []byte(tokenA+"|"+tokenB)...)
}
