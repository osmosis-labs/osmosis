package types

const (
	// ModuleName defines the module name.
	ModuleName   = "txfees"
	KeySeparator = "|"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey is the message route for slashing.
	RouterKey = ModuleName

	// NonNativeTxFeeCollectorName is the module account name for the alt fee collector account address (used for auto-swapping non-OSMO tx fees).
	// N.B. OSMO goes to authtypes.FeeCollectorName, matching the normal SDK flow.
	NonNativeTxFeeCollectorName = "non_native_fee_collector"

	// TakerFeeCommunityPoolName is the name of the module account that collects non-native taker fees, swaps, and sends them to the community pool.
	// Note, all taker fees initially get sent to the TakerFeeCollectorName, and then prior to the taker fees slated for the community pool being swapped and sent to the community pool, they are sent to this account.
	// This is done so that, in the event of a failed swap, the funds slated for the community pool are not grouped back with the rest of the taker fees in the next epoch.
	TakerFeeCommunityPoolName = "non_native_fee_collector_community_pool"

	// TakerFeeStakersName is the name of the module account that collects non-native taker fees, swaps, and sends them to the auth module account for stakers.
	// Note, all taker fees initially get sent to the TakerFeeCollectorName, and then prior to the taker fees slated for stakers being swapped and sent to stakers, they are sent to this account.
	// This is done so that, in the event of a failed swap, the funds slated for stakers are not grouped back with the rest of the taker fees in the next epoch.
	TakerFeeStakersName = "non_native_fee_collector_stakers"

	// TakerFeeBurnName is the name of the module account that collects non-native taker fees, swaps, and sends them to the burn address.
	// Note, all taker fees initially get sent to the TakerFeeCollectorName, and then prior to the taker fees slated for burning being swapped and sent to burn address, they are sent to this account.
	// This is done so that, in the event of a failed swap, the funds slated for burning are not grouped back with the rest of the taker fees in the next epoch.
	TakerFeeBurnName = "non_native_fee_collector_burn"

	// TakerFeeCollectorName is the module account name for the taker fee collector account address. It collects both native and non-native taker fees.
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
