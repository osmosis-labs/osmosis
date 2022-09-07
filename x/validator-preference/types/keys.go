package types

var (
	// ModuleName defines the module name
	ModuleName = "validator-set-preference"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing.
	RouterKey = ModuleName

	// KeyPrefixSuperfluidAsset defines prefix key for validator set.
	KeyPrefixValidatorSet = []byte{0x01}
)
