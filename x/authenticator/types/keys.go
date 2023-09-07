package types

import (
	"fmt"
	"strings"

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

	AttributeValueCategory        = ModuleName
	AttributeKeyAuthenticatorType = "authenticator_type"
)

var (
	KeyNextAccountAuthenticatorIdPrefix = []byte{0x01}
	KeyAccountAuthenticatorsPrefix      = []byte{0x02}

	// KeyAuthenticatorStoragePrefix is the prefix under which authenticators can keep their private state
	KeyAuthenticatorStorePrefix = []byte{0xFF}
)

// buildKey creates a key by concatenating the provided elements with the key separator.
func buildKey(elements ...interface{}) []byte {
	strElements := make([]string, len(elements))
	for i, element := range elements {
		strElements[i] = fmt.Sprint(element)
	}
	return []byte(strings.Join(strElements, KeySeparator))
}

func KeyAccount(account sdk.AccAddress) []byte {
	accBech32 := sdk.MustBech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, account)
	return buildKey(KeyAccountAuthenticatorsPrefix, accBech32)
}

func KeyAccountId(account sdk.AccAddress, id uint64) []byte {
	accBech32 := sdk.MustBech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, account)
	return buildKey(KeyAccountAuthenticatorsPrefix, accBech32, id)
}

func KeyNextAccountAuthenticatorId() []byte {
	return buildKey(KeyNextAccountAuthenticatorIdPrefix)
}

func KeyAuthenticatorStorage(authenticatorType string) []byte {
	// TODO: review for key malleability attacks. Maybe use a hash of the authenticator type?
	// Sholdn't matter that much since types are manually registered in code, and cosmwasm authenticator subtypes should
	// get a type based on contract addr. But still.
	return buildKey(KeyAuthenticatorStorePrefix, authenticatorType)
}
