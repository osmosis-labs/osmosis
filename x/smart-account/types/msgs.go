package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgAddAuthenticator    = "add_authenticator"
	TypeMsgRemoveAuthenticator = "remove_authenticator"
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

// MsgAddAuthenticator
var _ sdk.Msg = &MsgAddAuthenticator{}

func (msg *MsgAddAuthenticator) ValidateBasic() error {
	return validateSender(msg.Sender)
}

func (msg *MsgAddAuthenticator) GetSigners() []sdk.AccAddress {
	return getSender(msg.Sender)
}

func (msg MsgAddAuthenticator) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgAddAuthenticator) Route() string { return RouterKey }

func (msg MsgAddAuthenticator) Type() string { return TypeMsgAddAuthenticator }

// MsgRemoveAuthenticator
var _ sdk.Msg = &MsgRemoveAuthenticator{}

func (msg *MsgRemoveAuthenticator) ValidateBasic() error {
	return validateSender(msg.Sender)
}

func (msg MsgRemoveAuthenticator) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgRemoveAuthenticator) Route() string { return RouterKey }

func (msg MsgRemoveAuthenticator) Type() string { return TypeMsgRemoveAuthenticator }

func (msg *MsgRemoveAuthenticator) GetSigners() []sdk.AccAddress {
	return getSender(msg.Sender)
}

// MsgSetActiveState
var _ sdk.Msg = &MsgSetActiveState{}

func (msg *MsgSetActiveState) ValidateBasic() error {
	return validateSender(msg.Sender)
}

func (msg *MsgSetActiveState) GetSigners() []sdk.AccAddress {
	return getSender(msg.Sender)
}
