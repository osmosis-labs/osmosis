package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v21/x/authenticator/utils"
)

const (
	// ModuleName defines the module name
	ModuleName = "authenticator"

	// StoreKey defines the primary module store key
	ManagerStoreKey       = ModuleName + "manager"
	AuthenticatorStoreKey = ModuleName + "authenticator"

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	AttributeValueCategory        = ModuleName
	AttributeKeyAuthenticatorType = "authenticator_type"
)

var (
	KeyNextAccountAuthenticatorIdPrefix = []byte{0x01}
	KeyAccountAuthenticatorsPrefix      = []byte{0x02}
	KeyMaximumUnauthenticatedGas        = []byte("MaximumUnauthenticatedGas")
	KeyCosignerContract                 = []byte("CosignerContract")
)

func KeyAccount(account sdk.AccAddress) []byte {
	return utils.BuildKey(KeyAccountAuthenticatorsPrefix, account.String())
}

func KeyAccountId(account sdk.AccAddress, id uint64) []byte {
	return utils.BuildKey(KeyAccountAuthenticatorsPrefix, account.String(), id)
}

func KeyNextAccountAuthenticatorId() []byte {
	return utils.BuildKey(KeyNextAccountAuthenticatorIdPrefix)
}

func KeyAccountAuthenticatorsPrefixId() []byte {
	return utils.BuildKey(KeyAccountAuthenticatorsPrefix)
}
