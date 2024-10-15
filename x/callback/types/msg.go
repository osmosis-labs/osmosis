package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgRequestCallback{}
	_ sdk.Msg = &MsgCancelCallback{}
)

// NewMsgRequestCallback creates a new MsgRequestCallback instance.
func NewMsgRequestCallback(
	senderAddr sdk.AccAddress,
	contractAddr sdk.AccAddress,
	jobId uint64,
	callbackHeight int64,
	fees sdk.Coin,
) *MsgRequestCallback {
	msg := &MsgRequestCallback{
		Sender:          senderAddr.String(),
		ContractAddress: contractAddr.String(),
		JobId:           jobId,
		CallbackHeight:  callbackHeight,
		Fees:            fees,
	}

	return msg
}

// GetSigners implements the sdk.Msg interface.
func (m MsgRequestCallback) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgRequestCallback) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if _, err := sdk.AccAddressFromBech32(m.ContractAddress); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid contract address: %v", err)
	}

	return nil
}

// NewMsgCancelCallback creates a new MsgCancelCallback instance.
func NewMsgCancelCallback(
	senderAddr sdk.AccAddress,
	contractAddr sdk.AccAddress,
	jobId uint64,
	callbackHeight int64,
) *MsgCancelCallback {
	msg := &MsgCancelCallback{
		Sender:          senderAddr.String(),
		ContractAddress: contractAddr.String(),
		JobId:           jobId,
		CallbackHeight:  callbackHeight,
	}

	return msg
}

// GetSigners implements the sdk.Msg interface.
func (m MsgCancelCallback) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgCancelCallback) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if _, err := sdk.AccAddressFromBech32(m.ContractAddress); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid contract address: %v", err)
	}

	return nil
}
