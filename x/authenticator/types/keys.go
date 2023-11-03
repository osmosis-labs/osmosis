package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/x/authenticator/utils"
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
)

func KeyAccount(account sdk.AccAddress) []byte {
	accBech32 := sdk.MustBech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, account)
	return utils.BuildKey(KeyAccountAuthenticatorsPrefix, accBech32)
}

func KeyAccountId(account sdk.AccAddress, id uint64) []byte {
	accBech32 := sdk.MustBech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, account)
	return utils.BuildKey(KeyAccountAuthenticatorsPrefix, accBech32, id)
}

func KeyNextAccountAuthenticatorId() []byte {
	return utils.BuildKey(KeyNextAccountAuthenticatorIdPrefix)
}
