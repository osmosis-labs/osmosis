package types

import (
	fmt "fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "smartaccount"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	KeySeparator = "|"

	AttributeValueCategory        = ModuleName
	AttributeKeyAuthenticatorType = "authenticator_type"
	AttributeKeyAuthenticatorId   = "authenticator_id"

	AtrributeKeyIsSmartAccountActive = "is_smart_account_active"

	AttributeKeyAccountSequenceAuthenticator = "authenticator_acc_seq"
	AttributeKeySignatureAuthenticator       = "authenticator_signature"
)

var (
	// Store prefix keys
	KeyNextAccountAuthenticatorIdPrefix = []byte{0x01}
	KeyAccountAuthenticatorsPrefix      = []byte{0x02}

	// Parameter keys
	KeyMaximumUnauthenticatedGas = []byte("MaximumUnauthenticatedGas")
	KeyIsSmartAccountActive      = []byte("IsSmartAccountActive")
	KeyCircuitBreakerControllers = []byte("CircuitBreakerControllers")
)

func KeyAccount(account sdk.AccAddress) []byte {
	return BuildKey(KeyAccountAuthenticatorsPrefix, account.String())
}

func KeyAccountId(account sdk.AccAddress, id uint64) []byte {
	return BuildKey(KeyAccountAuthenticatorsPrefix, account.String(), id)
}

func KeyNextAccountAuthenticatorId() []byte {
	return BuildKey(KeyNextAccountAuthenticatorIdPrefix)
}

func KeyAccountAuthenticatorsPrefixId() []byte {
	return BuildKey(KeyAccountAuthenticatorsPrefix)
}

// BuildKey creates a key by concatenating the provided elements with the key separator.
func BuildKey(elements ...interface{}) []byte {
	strElements := make([]string, len(elements))
	for i, element := range elements {
		strElements[i] = fmt.Sprint(element)
	}
	return []byte(strings.Join(strElements, KeySeparator) + KeySeparator)
}
