package types

const (
	// ModuleName defines the module name.
	ModuleName   = "txfees"
	KeySeparator = "|"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey is the message route for slashing.
	RouterKey = ModuleName

	// NonNativeTxFeeCollectorName the module account name for the alt fee collector account address (used for auto-swapping non-OSMO tx fees).
	NonNativeTxFeeCollectorName = "non_native_fee_collector"

	// DeprecatedFeeCollectorForCommunityPoolName the module account name for the alt fee collector account address (used for auto-swapping non-OSMO tx fees).
	// These fees go to the community pool.
	// This module account is deprecated and we instead just send all taker fees to the taker fee collector, regardless of the denom.
	DeprecatedFeeCollectorForCommunityPoolName = "non_native_fee_collector_community_pool"

	// TakerFeeCollectorName the module account name for the taker fee collector account address. It collects both native and non-native taker fees.
	TakerFeeCollectorName = "taker_fee_collector"

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	BaseDenomKey                       = []byte("base_denom")
	FeeTokensStorePrefix               = []byte("fee_tokens")
	KeyTxFeeProtorevTracker            = []byte("txfee_protorev_tracker")
	KeyTxFeeProtorevTrackerStartHeight = []byte("txfee_protorev_tracker_start_height")
)
