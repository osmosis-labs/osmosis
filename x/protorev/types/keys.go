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
	prefixNumberOfTrades
	prefixProfitsByDenom
	prefixTradesByPool
	prefixProtoRevEnabled
	prefixAdminAccount
	prefixDeveloperAccount
	prefixDaysSinceGenesis
	prefixDeveloperFees
)

var (
	// -------------- Keys for trading stores -------------- //
	// KeyPrefixTokenPairRoutes is the prefix for the TokenPairArbRoutes store
	KeyPrefixTokenPairRoutes = []byte{prefixTokenPairRoutes}

	// KeyPrefixOsmoPools is the prefix for the osmo pool store
	KeyPrefixOsmoPools = []byte{prefixOsmoPools}

	// KeyPrefixAtomPools is the prefix for the atom pool store
	KeyPrefixAtomPools = []byte{prefixAtomPools}

	// -------------- Keys for statistics stores -------------- //
	// KeyPrefixNumberOfTrades is the prefix for store that keeps track of the number of trades executed
	KeyPrefixNumberOfTrades = []byte{prefixNumberOfTrades}

	// KeyPrefixProfitByDenom is the prefix for store that keeps track of the profits made by coin
	KeyPrefixProfitByDenom = []byte{prefixProfitsByDenom}

	// KeyPrefixTradesByPool is the prefix for store that keeps track of the trades made by pool
	KeyPrefixTradesByPool = []byte{prefixTradesByPool}

	// -------------- Keys for configuration/profit stores -------------- //
	// KeyPrefixProtoRevEnabled is the prefix for store that keeps track of whether protorev is enabled
	KeyPrefixProtoRevEnabled = []byte{prefixProtoRevEnabled}

	// KeyPrefixAdminAccount is the prefix for store that keeps track of the admin account
	KeyPrefixAdminAccount = []byte{prefixAdminAccount}

	// KeyPrefixDeveloperAccount is the prefix for store that keeps track of the developer account
	KeyPrefixDeveloperAccount = []byte{prefixDeveloperAccount}

	// KeyPrefixDaysSinceGenesis is the prefix for store that keeps track of the number of days since genesis
	KeyPrefixDaysSinceGenesis = []byte{prefixDaysSinceGenesis}

	// KeyPrefixDeveloperFees is the prefix for store that keeps track of the developer fees
	KeyPrefixDeveloperFees = []byte{prefixDeveloperFees}
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
	return append(KeyPrefixTokenPairRoutes, []byte(tokenA+tokenB)...)
}

// Returns the key needed to fetch the profit by coin
func GetKeyPrefixProfitByDenom(denom string) []byte {
	return append(KeyPrefixProfitByDenom, []byte(denom)...)
}

// Returns the key needed to fetch the number of trades by pool
func GetKeyPrefixTradesByPool(poolId uint64) []byte {
	poolIdBytes := UInt64ToBytes(poolId)
	return append(KeyPrefixTradesByPool, poolIdBytes...)
}

// Returns the key needed to fetch the developer fees by coin
func GetKeyPrefixDeveloperFees(denom string) []byte {
	return append(KeyPrefixDeveloperFees, []byte(denom)...)
}
