package types

import (
	"fmt"
	"strconv"
	"strings"
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
	prefixOsmoPools
	prefixAtomPools
	prefixNumberOfTrades
	prefixProfitsByDenom
	prefixTradesByRoute
	prefixProfitsByRoute
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
	// KeyPrefixNumberOfTrades is the prefix for the store that keeps track of the number of trades executed
	KeyPrefixNumberOfTrades = []byte{prefixNumberOfTrades}

	// KeyPrefixProfitByDenom is the prefix for the store that keeps track of the profits
	KeyPrefixProfitByDenom = []byte{prefixProfitsByDenom}

	// KeyPrefixTradesByRoute is the prefix for the store that keeps track of the number of trades executed by route
	KeyPrefixTradesByRoute = []byte{prefixTradesByRoute}

	// KeyPrefixProfitsByRoute is the prefix for the store that keeps track of the profits made by route
	KeyPrefixProfitsByRoute = []byte{prefixProfitsByRoute}
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
