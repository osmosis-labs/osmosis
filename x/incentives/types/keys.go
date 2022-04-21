package types

var (
	// ModuleName defines the module name.
	ModuleName = "incentives"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey is the message route for slashing.
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key.
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key.
	MemStoreKey = "mem_capability"

	// KeyPrefixTimestamp defines prefix key for timestamp iterator key.
	KeyPrefixTimestamp = []byte{0x01}

	// KeyLastGaugeID defines key for setting last gauge ID.
	KeyLastGaugeID = []byte{0x02}

	// KeyPrefixPeriodGauge defines prefix key for storing gauges.
	KeyPrefixPeriodGauge = []byte{0x03}

	// KeyPrefixGauges defines prefix key for storing reference key for all gauges.
	KeyPrefixGauges = []byte{0x04}

	// KeyPrefixUpcomingGauges defines prefix key for storing reference key for upcoming gauges.
	KeyPrefixUpcomingGauges = []byte{0x04, 0x00}

	// KeyPrefixActiveGauges defines prefix key for storing reference key for active gauges.
	KeyPrefixActiveGauges = []byte{0x04, 0x01}

	// KeyPrefixFinishedGauges defines prefix key for storing reference key for finished gauges.
	KeyPrefixFinishedGauges = []byte{0x04, 0x02}

	// KeyPrefixGaugesByDenom defines prefix key for storing indexes of gauge IDs by denomination.
	KeyPrefixGaugesByDenom = []byte{0x05}

	// KeyIndexSeparator defines key for merging bytes.
	KeyIndexSeparator = []byte{0x07}

	// LockableDurationsKey defines key for storing valid durations for giving incentives.
	LockableDurationsKey = []byte("lockable_durations")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
