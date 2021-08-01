package types

const (
	// ModuleName defines the module name
	ModuleName = "txfees"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_capability"

	// this line is used by starport scaffolding # ibc/keys/name
)

var (
	BaseDenomKey         = []byte("base_denom")
	FeeTokensStorePrefix = []byte("fee_tokens")
)
