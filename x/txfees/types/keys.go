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

	// NonNativeFeeCollectorForStakingRewardsName the module account name for the alt fee collector account address (used for auto-swapping non-OSMO tx fees).
	// These fees go to the staking rewards pool.
	NonNativeFeeCollectorForStakingRewardsName = "non_native_fee_collector"

	// NonNativeFeeCollectorForCommunityPoolName the module account name for the alt fee collector account address (used for auto-swapping non-OSMO tx fees).
	// These fees go to the community pool.
	NonNativeFeeCollectorForCommunityPoolName = "non_native_fee_collector_community_pool"

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	BaseDenomKey         = []byte("base_denom")
	FeeTokensStorePrefix = []byte("fee_tokens")
)
