package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// constants.
const (
	TypeMsgEmitIBCAck = "emit-ibc-ack"
)

var _ sdk.Msg = &MsgEmitIBCAck{}

func (m MsgEmitIBCAck) Route() string { return RouterKey }
func (m MsgEmitIBCAck) Type() string  { return TypeMsgEmitIBCAck }
func (m MsgEmitIBCAck) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}
	return nil
}

func (m MsgEmitIBCAck) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}
