package proto

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgExitLBP{}

func (msg *MsgExitLBP) ValidateBasic() error {
	return nil
	// if !msg.Amount.IsPositive() {
	// 	return errors.Wrap(errors.ErrInvalidCoins, "amount must be a positive integer")
	// }
}

func (msg *MsgExitLBP) GetSigners() []sdk.AccAddress {
	a, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{a}
}

// TODO: remove when updating to SDK v0.44+
// Deprecated methods

func (msg *MsgExitLBP) GetSignBytes() []byte {
	panic("not implemented")
}

func (msg *MsgExitLBP) Route() string {
	panic("not implemented")
}

func (msg *MsgExitLBP) Type() string {
	panic("not implemented")
}
