package types

import (
	errorsmod "cosmossdk.io/errors"
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateCronJob = "update_crone_job"

var _ sdk.Msg = &MsgUpdateCronJob{}

func NewMsgUpdateCronJob(
	securityAddress string,
	cronID uint64,
	contractAddress string,
	jsonMsg string,
) *MsgUpdateCronJob {
	return &MsgUpdateCronJob{
		SecurityAddress: securityAddress,
		Id:              cronID,
		ContractAddress: contractAddress,
		JsonMsg:         jsonMsg,
	}
}

func (msg *MsgUpdateCronJob) Route() string {
	return RouterKey
}

func (msg *MsgUpdateCronJob) Type() string {
	return TypeMsgUpdateCronJob
}

func (msg *MsgUpdateCronJob) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.SecurityAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateCronJob) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateCronJob) ValidateBasic() error {
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
	if len(msg.JsonMsg) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "json msg cannot be empty")
	}
	if !json.Valid([]byte(msg.JsonMsg)) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "json msg is invalid")
	}
	return nil
}
