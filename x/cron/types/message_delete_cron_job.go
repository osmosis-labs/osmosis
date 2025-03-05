package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDeleteCronJob = "delete_crone_job"

var _ sdk.Msg = &MsgDeleteCronJob{}

func NewMsgDeleteCronJob(
	securityAddress string,
	cronID uint64,
	contractAddress string,
) *MsgDeleteCronJob {
	return &MsgDeleteCronJob{
		SecurityAddress: securityAddress,
		Id:              cronID,
		ContractAddress: contractAddress,
	}
}

func (msg *MsgDeleteCronJob) Route() string {
	return RouterKey
}

func (msg *MsgDeleteCronJob) Type() string {
	return TypeMsgDeleteCronJob
}

func (msg *MsgDeleteCronJob) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.SecurityAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteCronJob) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteCronJob) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.SecurityAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid security address (%s)", err)
	}
	if msg.Id == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cron id cannot be 0")
	}
	if len(msg.ContractAddress) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "contract address cannot be empty")
	}
	_, err = sdk.AccAddressFromBech32(msg.ContractAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", err)
	}
	return nil
}
