package api

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgFinalizeLBP{}

func (msg *MsgFinalizeLBP) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errors.ErrInvalidRequest.Wrapf("Invalid sender address (%s)", err)
	}
	return nil
}

func (msg *MsgFinalizeLBP) Validate(now, end time.Time) error {
	if now.Before(end) {
		return errors.ErrInvalidRequest.Wrapf("You can exit the LBP only once the LBP ended (%s)", end)
	}
	return nil
}

func (msg *MsgFinalizeLBP) GetSigners() []sdk.AccAddress {
	a, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{a}
}
