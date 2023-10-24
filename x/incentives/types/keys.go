package types

import "fmt"

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

	// KeyPrefixGroup defines prefix key for storing groups.
	KeyPrefixGroup = []byte{0x08}

	// LockableDurationsKey defines key for storing valid durations for giving incentives.
	LockableDurationsKey = []byte("lockable_durations")

	NoLockInternalPrefix = "no-lock/i/"
	NoLockExternalPrefix = "no-lock/e/"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// NoLockExternalGaugeDenom returns the gauge denom for the no-lock external gauge for the given pool ID.
func NoLockExternalGaugeDenom(poolId uint64) string {
	return fmt.Sprintf("%s%d", NoLockExternalPrefix, poolId)
}

// NoLockInternalGaugeDenom returns the gauge denom for the no-lock internal gauge for the given pool ID.
func NoLockInternalGaugeDenom(poolId uint64) string {
	return fmt.Sprintf("%s%d", NoLockInternalPrefix, poolId)
}

// KeyGroupByGaugeID returns group key for a given groupGaugeId.
func KeyGroupByGaugeID(groupGaugeId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d%s", KeyPrefixGroup, groupGaugeId, KeyIndexSeparator))
}
