package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "authenticator"

	// StoreKey defines the primary module store key
	StoreKey     = ModuleName
	KeySeparator = "|"

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_authenticator"
)

var (
	// KeyAuthenticators
	KeyAuthenticators = []byte{0x01}
)

func KeyAccount(account sdk.AccAddress) []byte {
	// ToDo: Do we want to encode the authenticator id in the key and use two different prefixes??
	accBech32 := sdk.MustBech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, account)
	return []byte(fmt.Sprintf("%s%s", accBech32, KeySeparator))
}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
