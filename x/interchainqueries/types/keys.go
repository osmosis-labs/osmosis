package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "interchainqueries"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_interchainqueries"
)

const (
	prefixRegisteredQuery = iota + 1
	prefixRegisteredQueryResult
	prefixSubmittedTx
	prefixTxQueryToRemove
)

var (
	// RegisteredQueryKey is the store key for queries registered in the module.
	RegisteredQueryKey = []byte{prefixRegisteredQuery}
	// RegisteredQueryResultKey is the store key for KV query results.
	RegisteredQueryResultKey = []byte{prefixRegisteredQueryResult}
	// SubmittedTxKey is the store key for submitted transaction hashes.
	SubmittedTxKey = []byte{prefixSubmittedTx}
	// TxQueryToRemoveKey is the store key for TX queries marked to be removed.
	TxQueryToRemoveKey = []byte{prefixTxQueryToRemove}
	// LastRegisteredQueryIDKey is the store key for last registered query ID.
	LastRegisteredQueryIDKey = []byte{0x64}
)

// GetRegisteredQueryByIDKey builds a store key to access a registered query by query ID.
func GetRegisteredQueryByIDKey(id uint64) []byte {
	return append(RegisteredQueryKey, sdk.Uint64ToBigEndian(id)...)
}

// GetSubmittedTransactionIDForQueryKeyPrefix builds a store key prefix to access TX query hashes by ID.
func GetSubmittedTransactionIDForQueryKeyPrefix(queryID uint64) []byte {
	return append(SubmittedTxKey, sdk.Uint64ToBigEndian(queryID)...)
}

// GetSubmittedTransactionIDForQueryKey builds a store key to access a submitted transaction hash
// by query ID and hash.
func GetSubmittedTransactionIDForQueryKey(queryID uint64, txHash []byte) []byte {
	return append(GetSubmittedTransactionIDForQueryKeyPrefix(queryID), txHash...)
}

// GetRegisteredQueryResultByIDKey builds a store key to access a KV query result by query ID.
func GetRegisteredQueryResultByIDKey(id uint64) []byte {
	return append(RegisteredQueryResultKey, sdk.Uint64ToBigEndian(id)...)
}

// GetTxQueryToRemoveByIDKey builds a store key to access a TX query marked to be removed.
func GetTxQueryToRemoveByIDKey(id uint64) []byte {
	return append(TxQueryToRemoveKey, sdk.Uint64ToBigEndian(id)...)
}
