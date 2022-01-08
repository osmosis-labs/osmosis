package api

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgWithdraw{}

func (msg *MsgWithdraw) ValidateBasic() error {
	return nil
	// if !msg.Amount.IsPositive() {
	// 	return errors.Wrap(errors.ErrInvalidCoins, "amount must be a positive integer")
	// }
}

func (msg *MsgWithdraw) GetSigners() []sdk.AccAddress {
	a, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{a}
}
