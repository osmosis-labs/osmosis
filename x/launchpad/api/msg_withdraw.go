package api

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgWithdraw{}

func (msg *MsgWithdraw) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	return errors.ErrInvalidRequest.Wrapf("Invalid sender address (%s)", err)
}

func (msg *MsgWithdraw) Validate() error {
	if msg.Amount != nil && !msg.Amount.IsPositive() {
		return errors.ErrInvalidCoins.Wrap("amount must be a positive integer")
	}
	return nil
}

func (msg *MsgWithdraw) GetSigners() []sdk.AccAddress {
	a, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{a}
}
