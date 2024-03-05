package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgInboundTransfer{}

// NewMsgInboundTransfer creates a message to transfer tokens from the source chain
func NewMsgInboundTransfer(
	sender string,
	destination string,
	asset Asset,
	amount sdk.DecCoin,
) *MsgInboundTransfer {
	return &MsgInboundTransfer{
		Sender:      sender,
		Destination: destination,
		Asset:       asset,
		Amount:      amount,
	}
}

func (m MsgInboundTransfer) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(m.Destination)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid destination address (%s)", err)
	}

	err = m.Asset.Validate()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAsset, err.Error())
	}

	if !m.Amount.IsZero() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

func (m MsgInboundTransfer) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgOutboundTransfer{}

// NewMsgInboundTransfer creates a message to transfer tokens from the source chain
func NewMsgOutboundTransfer(
	sender string,
	destination string,
	asset Asset,
	amount sdk.Coin,
) *MsgOutboundTransfer {
	return &MsgOutboundTransfer{
		Sender:      sender,
		Destination: destination,
		Asset:       asset,
		Amount:      amount,
	}
}

func (m MsgOutboundTransfer) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(m.Destination)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid destination address (%s)", err)
	}

	err = m.Asset.Validate()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAsset, err.Error())
	}

	if !m.Amount.IsZero() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

func (m MsgOutboundTransfer) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

func NewMsgEnableBridge(
	sender string,
	asset Asset,
) *MsgEnableBridge {
	return &MsgEnableBridge{
		Sender: sender,
		Asset:  asset,
	}
}

func (m MsgEnableBridge) ValidateBasic() error {
	err := m.Asset.Validate()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAsset, err.Error())
	}
	return nil
}

func (m MsgEnableBridge) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

func NewMsgDisableBridge(
	sender string,
	asset Asset,
) *MsgDisableBridge {
	return &MsgDisableBridge{
		Sender: sender,
		Asset:  asset,
	}
}

func (m MsgDisableBridge) ValidateBasic() error {
	err := m.Asset.Validate()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAsset, err.Error())
	}
	return nil
}

func (m MsgDisableBridge) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}
