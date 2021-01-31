package types

var (
	// ModuleName defines the module name
	ModuleName = "lockup"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// KeyLastLockID defines key to store lock ID used by last
	KeyLastLockID = []byte{0x33}

	// KeyPrefixTimestamp defines prefix key for timestamp iterator key
	KeyPrefixTimestamp = []byte{0x34}

	// KeyPrefixDuration defines prefix key for duration iterator key
	KeyPrefixDuration = []byte{0x35}

	// KeyPrefixLockDuration defines prefix for the iteration of lock IDs by duration
	KeyPrefixLockDuration = []byte{0x36}

	// KeyPrefixAccountLockDuration defines prefix for the iteration of lock IDs by account and duration
	KeyPrefixAccountLockDuration = []byte{0x37}

	// KeyPrefixDenomLockDuration defines prefix for the iteration of lock IDs by denom and duration
	KeyPrefixDenomLockDuration = []byte{0x38}

	// KeyPrefixAccountDenomLockDuration defines prefix for the iteration of lock IDs by account, denomination and duration
	KeyPrefixAccountDenomLockDuration = []byte{0x39}

	// KeyPrefixPeriodLock defines prefix to store period lock by ID
	KeyPrefixPeriodLock = []byte{0x00}

	// KeyPrefixLockTimestamp defines prefix for the iteration of lock IDs by timestamp
	KeyPrefixLockTimestamp = []byte{0x01}

	// KeyPrefixAccountLockTimestamp defines prefix for the iteration of lock IDs by account and timestamp
	KeyPrefixAccountLockTimestamp = []byte{0x02}

	// KeyPrefixDenomLockTimestamp defines prefix for the iteration of lock IDs by denom and timestamp
	KeyPrefixDenomLockTimestamp = []byte{0x03}

	// KeyPrefixAccountDenomLockTimestamp defines prefix for the iteration of lock IDs by account, denomination and timestamp
	KeyPrefixAccountDenomLockTimestamp = []byte{0x04}

	// KeyIndexSeparator defines separator between keys when combine
	KeyIndexSeparator = []byte{0x7f}
)
