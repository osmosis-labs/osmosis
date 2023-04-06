package types

var (
	// ModuleName defines the module name
	ModuleName = "valsetpref"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// KeyPrefixValidatorSet defines prefix key for validator set.
	KeyPrefixValidatorSet = []byte{0x01}

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)
