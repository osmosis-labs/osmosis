package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgToggleCronJob = "toggle_crone_job"

var _ sdk.Msg = &MsgToggleCronJob{}

func NewMsgToggleCronJob(
	securityAddress string,
	cronID uint64,
) *MsgToggleCronJob {
	return &MsgToggleCronJob{
		SecurityAddress: securityAddress,
		Id:              cronID,
	}
}

func (msg *MsgToggleCronJob) Route() string {
	return RouterKey
}

func (msg *MsgToggleCronJob) Type() string {
	return TypeMsgToggleCronJob
}

func (msg *MsgToggleCronJob) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.SecurityAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgToggleCronJob) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgToggleCronJob) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.SecurityAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid security address (%s)", err)
	}
	if msg.Id == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cron id cannot be 0")
	}
	return nil
}
