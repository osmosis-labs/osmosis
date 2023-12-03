package types

const (
	// ModuleName defines the module name.
	ModuleName = "txfees"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey is the message route for slashing.
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// FeeCollectorName the module account name for the fee collector account address.
	FeeCollectorName = "fee_collector"
)

var (
	BaseDenomKey         = []byte("base_denom")
	FeeTokensStorePrefix = []byte("fee_tokens")
)
