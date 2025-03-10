package types

import (
	errorsmod "cosmossdk.io/errors"
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRegisterCron = "register_cron"

var _ sdk.Msg = &MsgRegisterCron{}

func NewMsgRegisterCron(
	securityAddress string,
	name string,
	description string,
	contractAddress string,
	jsonMsg string,
) *MsgRegisterCron {
	return &MsgRegisterCron{
		SecurityAddress: securityAddress,
		Name:            name,
		ContractAddress: contractAddress,
		Description:     description,
		JsonMsg:         jsonMsg,
	}
}

func (msg *MsgRegisterCron) Route() string {
	return RouterKey
}

func (msg *MsgRegisterCron) Type() string {
	return TypeMsgRegisterCron
}

func (msg *MsgRegisterCron) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.SecurityAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRegisterCron) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRegisterCron) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.SecurityAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.ContractAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address (%s)", err)
	}

	if len(msg.Name) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "name cannot be empty")
	}

	if len(msg.Name) > 20 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "name cannot be longer than 20 characters")
	}

	if len(msg.Description) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "description cannot be empty")
	}

	if len(msg.Description) > 100 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "description cannot be longer than 100 characters")
	}

	if len(msg.JsonMsg) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "json msg cannot be empty")
	}

	if !json.Valid([]byte(msg.JsonMsg)) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "json msg is not a valid json")
	}
	return nil
}
