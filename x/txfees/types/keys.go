package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "txfees"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// FeeCollectorName the root string for the fee collector account address
	FeeCollectorName = "fee_collector"

	// FooCollectorName the root string for the $FOO collector account address (used for auto-swapping non-OSMO tx fees)
	FooCollectorName = "foo_collector"

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	BaseDenomKey         = []byte("base_denom")
	FeeTokensStorePrefix = []byte("fee_tokens")

	// AddressStoreKeyPrefix prefix for account-by-address store
	AddressStoreKeyPrefix = []byte{0x01}

	// param key for global account number
	GlobalAccountNumberKey = []byte("globalAccountNumber")
)

// AddressStoreKey turn an address to key used to get it from the account store
func AddressStoreKey(addr sdk.AccAddress) []byte {
	return append(AddressStoreKeyPrefix, addr.Bytes()...)
}