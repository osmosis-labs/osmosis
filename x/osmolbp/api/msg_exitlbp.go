package api

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgExitLBP{}

func (msg *MsgExitLBP) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	return errors.ErrInvalidRequest.Wrapf("Invalid sender address (%s)", err)
}

func (msg *MsgExitLBP) GetSigners() []sdk.AccAddress {
	a, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{a}
}
