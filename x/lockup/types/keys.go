package types

var (
	// ModuleName defines the module name.
	ModuleName = "lockup"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey is the message route for slashing.
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key.
	QuerierRoute = ModuleName

	// KeyLastLockID defines key to store lock ID used by last.
	KeyLastLockID = []byte{0x01}

	// KeyPrefixPeriodLock defines prefix to store period lock by ID.
	KeyPrefixPeriodLock = []byte{0x02}

	// KeyPrefixNotUnlocking defines prefix to query iterators which hasn't started unlocking.
	KeyPrefixNotUnlocking = []byte{0x03}

	// KeyPrefixUnlocking defines prefix to query iterators which has started unlocking.
	KeyPrefixUnlocking = []byte{0x04}

	// KeyPrefixTimestamp defines prefix key for timestamp iterator key.
	KeyPrefixTimestamp = []byte{0x05}

	// KeyPrefixDuration defines prefix key for duration iterator key.
	KeyPrefixDuration = []byte{0x06}

	// KeyPrefixLockDuration defines prefix for the iteration of lock IDs by duration.
	KeyPrefixLockDuration = []byte{0x07}

	// KeyPrefixAccountLockDuration defines prefix for the iteration of lock IDs by account and duration.
	KeyPrefixAccountLockDuration = []byte{0x08}

	// KeyPrefixDenomLockDuration defines prefix for the iteration of lock IDs by denom and duration.
	KeyPrefixDenomLockDuration = []byte{0x09}

	// KeyPrefixAccountDenomLockDuration defines prefix for the iteration of lock IDs by account, denomination and duration.
	KeyPrefixAccountDenomLockDuration = []byte{0x0A}

	// KeyPrefixLockTimestamp defines prefix for the iteration of lock IDs by timestamp.
	KeyPrefixLockTimestamp = []byte{0x0B}

	// KeyPrefixAccountLockTimestamp defines prefix for the iteration of lock IDs by account and timestamp.
	KeyPrefixAccountLockTimestamp = []byte{0x0C}

	// KeyPrefixDenomLockTimestamp defines prefix for the iteration of lock IDs by denom and timestamp.
	KeyPrefixDenomLockTimestamp = []byte{0x0D}

	// KeyPrefixAccountDenomLockTimestamp defines prefix for the iteration of lock IDs by account, denomination and timestamp.
	KeyPrefixAccountDenomLockTimestamp = []byte{0x0E}

	// KeyPrefixPeriodLock defines prefix to store synthetic lockup by lock ID and synthetic suffix.
	KeyPrefixSyntheticLockup = []byte{0x0F}

	// KeyPrefixSyntheticLockTimestamp defines prefix for the iteration of synthetic lockups by timestamp.
	KeyPrefixSyntheticLockTimestamp = []byte{0x10}

	// KeyPrefixLockAccumulation defines prefix for the lock accumulation store.
	KeyPrefixLockAccumulation = []byte{0x20}

	// KeyIndexSeparator defines separator between keys when combine, it should be one that is not used in denom expression.
	KeyIndexSeparator = []byte{0xFF}
)
