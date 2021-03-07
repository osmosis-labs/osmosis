package types

var (
	// ModuleName defines the module name
	ModuleName = "incentives"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_capability"

	// KeyPrefixTimestamp defines prefix key for timestamp iterator key
	KeyPrefixTimestamp = []byte{0x01}

	// KeyLastPotID defines key for setting last pot ID
	KeyLastPotID = []byte{0x02}

	// KeyPrefixPeriodPot defines prefix key for storing pots
	KeyPrefixPeriodPot = []byte{0x03}

	// KeyIndexSeparator defines key for merging bytes
	KeyIndexSeparator = []byte{0x7F}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
