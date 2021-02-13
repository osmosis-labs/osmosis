package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// bank message types
const (
	TypeMsgClaim = "send"
)

var _ sdk.Msg = &MsgClaim{}

// NewMsgClaim - construct a msg to claim coins from this module
func NewMsgClaim(sender sdk.AccAddress) *MsgClaim {
	return &MsgClaim{Sender: sender.String()}
}

// Route Implements Msg.
func (msg MsgClaim) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgClaim) Type() string { return TypeMsgClaim }

// ValidateBasic Implements Msg.
func (msg MsgClaim) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners Implements Msg.
func (msg MsgClaim) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}
