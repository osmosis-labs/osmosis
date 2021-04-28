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

	// KeyPrefixPots defines prefix key for storing reference key for all pots
	KeyPrefixPots = []byte{0x04}

	// KeyPrefixUpcomingPots defines prefix key for storing reference key for upcoming pots
	KeyPrefixUpcomingPots = []byte{0x04, 0x00}

	// KeyPrefixActivePots defines prefix key for storing reference key for active pots
	KeyPrefixActivePots = []byte{0x04, 0x01}

	// KeyPrefixFinishedPots defines prefix key for storing reference key for finished pots
	KeyPrefixFinishedPots = []byte{0x04, 0x02}

	// KeyIndexSeparator defines key for merging bytes
	KeyIndexSeparator = []byte{0x07}

	// KeyCurrentEpoch defines key for storing current epoch
	KeyCurrentEpoch = []byte{0x08}

	// KeyEpochBeginBlock defines key for storing begin block of current epoch
	KeyEpochBeginBlock = []byte{0x09}

	// KeyTotalLockedDenom defines key for storing total locked token amount per
	// denom
	KeyPrefixTotalLockedDenom = []byte{0x10}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
