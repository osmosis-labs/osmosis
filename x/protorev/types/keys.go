package types

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
	prefixDenomPairToPool
	prefixBaseDenoms
	prefixNumberOfTrades
	prefixProfitsByDenom
	prefixTradesByRoute
	prefixProfitsByRoute
	prefixDeveloperAccount
	prefixDaysSinceGenesis
	prefixDeveloperFees
	prefixMaxPoolPointsPerTx
	prefixMaxPoolPointsPerBlock
	prefixPoolPointCountForBlock
	prefixLatestBlockHeight
	prefixPoolWeights
)

var (
	// -------------- Keys for trading stores -------------- //
	// KeyPrefixTokenPairRoutes is the prefix for the TokenPairArbRoutes store
	KeyPrefixTokenPairRoutes = []byte{prefixTokenPairRoutes}

	// KeyPrefixDenomPairToPool is the prefix that is used to store the pool id for a given denom pair (baseDenom, otherDenom)
	KeyPrefixDenomPairToPool = []byte{prefixDenomPairToPool}

	// KeyPrefixBaseDenoms is the prefix that is used to store the base denoms that are used to create cyclic arbitrage routes
	KeyPrefixBaseDenoms = []byte{prefixBaseDenoms}

	// -------------- Keys for statistics stores -------------- //
	// KeyPrefixNumberOfTrades is the prefix for the store that keeps track of the number of trades executed
	KeyPrefixNumberOfTrades = []byte{prefixNumberOfTrades}

	// KeyPrefixProfitByDenom is the prefix for the store that keeps track of the profits
	KeyPrefixProfitByDenom = []byte{prefixProfitsByDenom}

	// KeyPrefixTradesByRoute is the prefix for the store that keeps track of the number of trades executed by route
	KeyPrefixTradesByRoute = []byte{prefixTradesByRoute}

	// KeyPrefixProfitsByRoute is the prefix for the store that keeps track of the profits made by route
	KeyPrefixProfitsByRoute = []byte{prefixProfitsByRoute}

	// -------------- Keys for configuration/admin stores -------------- //
	// KeyPrefixDeveloperAccount is the prefix for store that keeps track of the developer account
	KeyPrefixDeveloperAccount = []byte{prefixDeveloperAccount}

	// KeyPrefixDaysSinceGenesis is the prefix for store that keeps track of the number of days since genesis
	KeyPrefixDaysSinceGenesis = []byte{prefixDaysSinceGenesis}

	// KeyPrefixDeveloperFees is the prefix for store that keeps track of the developer fees
	KeyPrefixDeveloperFees = []byte{prefixDeveloperFees}

	// KeyPrefixMaxPointsPerTx is the prefix for store that keeps track of the max number of pool points that can be consumed per tx
	KeyPrefixMaxPointsPerTx = []byte{prefixMaxPoolPointsPerTx}

	// KeyPrefixMaxPointsPerBlock is the prefix for store that keeps track of the max number of pool points that can be consumed per block
	KeyPrefixMaxPointsPerBlock = []byte{prefixMaxPoolPointsPerBlock}

	// KeyPrefixPointCountForBlock is the prefix for store that keeps track of the number of pool points that have been consumed in the current block
	KeyPrefixPointCountForBlock = []byte{prefixPoolPointCountForBlock}

	// KeyPrefixLatestBlockHeight is the prefix for store that keeps track of the latest recorded block height
	KeyPrefixLatestBlockHeight = []byte{prefixLatestBlockHeight}

	// KeyPrefixPoolWeights is the prefix for store that keeps track of the weights for different pool types
	KeyPrefixPoolWeights = []byte{prefixPoolWeights}
)

// Returns the key needed to fetch the pool id for a given denom
func GetKeyPrefixDenomPairToPool(baseDenom, matchDenom string) []byte {
	return append(KeyPrefixDenomPairToPool, []byte(baseDenom+"|"+matchDenom)...)
}

// Returns the key needed to fetch info about base denoms
func GetKeyPrefixBaseDenom(priority uint64) []byte {
	return append(KeyPrefixBaseDenoms, sdk.Uint64ToBigEndian(priority)...)
}

// Returns the key needed to fetch the tokenPair routes for a given pair of tokens
func GetKeyPrefixRouteForTokenPair(tokenA, tokenB string) []byte {
	return append(KeyPrefixTokenPairRoutes, []byte(tokenA+"|"+tokenB)...)
}

// Returns the key needed to fetch the profit by coin
func GetKeyPrefixProfitByDenom(denom string) []byte {
	return append(KeyPrefixProfitByDenom, []byte(denom)...)
}

// Returns the key needed to fetch the number of trades by route
func GetKeyPrefixTradesByRoute(route []uint64) []byte {
	return append(KeyPrefixTradesByRoute, CreateRouteKey(route)...)
}

// Returns the key needed to fetch the profits by route
func GetKeyPrefixProfitsByRoute(route []uint64, denom string) []byte {
	return append(append(KeyPrefixProfitsByRoute, CreateRouteKey(route)...), []byte(denom)...)
}

// createRouteKey creates a key for the given route. converts a slice of uint64 to a string separated by a pipe
// {1,2,3,4} -> []byte("1|2|3|4")
func CreateRouteKey(route []uint64) []byte {
	return []byte(strings.Trim(strings.Join(strings.Fields(fmt.Sprint(route)), "|"), "[]"))
}

// createRouteFromKey creates a route from a key. converts a string separated by a pipe to a slice of uint64
// []byte("1|2|3|4") -> {1,2,3,4}
func CreateRouteFromKey(key []byte) ([]uint64, error) {
	var route []uint64
	for _, r := range strings.Split(string(key), "|") {
		pool, err := strconv.ParseUint(r, 10, 64)
		if err != nil {
			return []uint64{}, err
		}

		route = append(route, pool)
	}
	return route, nil
}

// Returns the key needed to fetch the developer fees by coin
func GetKeyPrefixDeveloperFees(denom string) []byte {
	return append(KeyPrefixDeveloperFees, []byte(denom)...)
}
