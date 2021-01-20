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

	// KeyPrefixPeriodLock defines history of PubKey history of an account
	KeyPrefixPeriodLock = []byte{0x01} // prefix for the timestamps of period lock
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
