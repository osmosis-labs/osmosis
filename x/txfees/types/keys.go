package types

const (
	// ModuleName defines the module name.
	ModuleName = "txfees"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey is the message route for slashing.
	RouterKey = ModuleName

	// FeeCollectorName the module account name for the fee collector account address.
	FeeCollectorName = "fee_collector"

	// NonNativeFeeCollectorName the module account name for the alt fee collector account address (used for auto-swapping non-OSMO tx fees).
	NonNativeFeeCollectorName = "non_native_fee_collector"

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	BaseDenomKey         = []byte("base_denom")
	FeeTokensStorePrefix = []byte("fee_tokens")
)
