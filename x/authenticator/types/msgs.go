package types

import (
	"fmt"

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
	return validateSender(msg.Sender)
}

func (msg *MsgAddAuthenticator) GetSigners() []sdk.AccAddress {
	senders := getSender(msg.Sender)

	if msg.Cosigner != "" {
		cosigner, err := sdk.AccAddressFromBech32(msg.Cosigner)
		if err != nil {
			panic(err)
		}
		senders = append(senders, cosigner)
	}
	return senders
}

var _ sdk.Msg = &MsgRemoveAuthenticator{}

func (msg *MsgRemoveAuthenticator) ValidateBasic() error {
	// TODO: call validate here
	return validateSender(msg.Sender)
}

func (msg *MsgRemoveAuthenticator) GetSigners() []sdk.AccAddress {
	return getSender(msg.Sender)
}
