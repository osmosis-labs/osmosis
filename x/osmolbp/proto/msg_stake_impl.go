package proto

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (msg *MsgStake) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %s", err)
	}
	return nil
}

func (msg *MsgStake) GetSigners() []sdk.AccAddress {
	a, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{a}
}

// TODO: remove when updating to SDK v0.44+
// Deprecated methods

func (msg *MsgStake) GetSignBytes() []byte {
	panic("not implemented")
}

func (msg *MsgStake) Route() string {
	panic("not implemented")
}

func (msg *MsgStake) Type() string {
	panic("not implemented")
}
