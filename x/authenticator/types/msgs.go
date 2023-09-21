package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Helper functions
func validateSender(sender string) error {
	_, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return fmt.Errorf("invalid sender address (%s)", err)
	}
	return nil
}

func getSender(sender string) []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// Msgs
var _ sdk.Msg = &MsgAddAuthenticator{}

func (msg *MsgAddAuthenticator) ValidateBasic() error {
	if len(msg.Data) != secp256k1.PubKeySize {
		return fmt.Errorf("invalid secp256k1 pub key size")
	}

	return validateSender(msg.Sender)
}

func (msg *MsgAddAuthenticator) GetSigners() []sdk.AccAddress {
	return getSender(msg.Sender)
}

var _ sdk.Msg = &MsgRemoveAuthenticator{}

func (msg *MsgRemoveAuthenticator) ValidateBasic() error {
	// TODO: call validate here
	return validateSender(msg.Sender)
}

func (msg *MsgRemoveAuthenticator) GetSigners() []sdk.AccAddress {
	return getSender(msg.Sender)
}
