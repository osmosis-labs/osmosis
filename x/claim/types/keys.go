package types

const (
	// ModuleName defines the module name
	ModuleName = "claim"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// ClaimableStoreKey defines the store key for claimable amounts
	ClaimableStoreKey = "claimable"

	// ParamsKey defines the store key for claim module parameters
	ParamsKey = "params"
)
