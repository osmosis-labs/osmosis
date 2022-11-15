package types

import (
	"fmt"
)

const (
	// ModuleName defines the module name
	ModuleName = "protorev"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// TransientStoreKey defines the module's transient store key
	TransientStoreKey = "transient_protorev"

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_protorev"
)

const (
	prefixNeedToArb = iota + 1
	prefixArbDetails
	prefixConnectedTokens
	prefixConnectedTokensToPoolIDs
	prefixPoolToRoutes
)

var (
	// ProtoRev Code
	KeyNeedToArb                = []byte{prefixNeedToArb}
	KeyArbDetails               = []byte{prefixArbDetails}
	KeyConnectedTokens          = []byte{prefixConnectedTokens}
	KeyConnectedTokensToPoolIDs = []byte{prefixConnectedTokensToPoolIDs}
	KeyPoolToRoutes             = []byte{prefixPoolToRoutes}
)

func GetConnectedTokensStoreKey(token *string) []byte {
	return []byte(fmt.Sprintf("token/%s", *token))
}

func GetConnectedTokensToPoolIDsStoreKey(tokenA, tokenB string) []byte {
	// Compare tokenA and tokenB to see which one is alphabetically first
	if tokenA < tokenB {
		return []byte(fmt.Sprintf("connected_tokens/%s/%s", tokenA, tokenB))
	} else {
		return []byte(fmt.Sprintf("connected_tokens/%s/%s", tokenB, tokenA))
	}
}

func GetPoolToRoutesStoreKey(poolId uint64) []byte {
	return []byte(fmt.Sprintf("pool_routes/%d", poolId))
}
