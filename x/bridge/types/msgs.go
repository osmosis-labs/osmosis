package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgInboundTransfer{}

func NewMsgInboundTransfer(
	sender string,
	destAddr string,
	asset Asset,
	amount math.Int,
) *MsgInboundTransfer {
	return &MsgInboundTransfer{
		Sender:   sender,
		DestAddr: destAddr,
		Asset:    asset,
		Amount:   amount,
	}
}

func (m MsgInboundTransfer) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(m.DestAddr)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid destination address (%s)", err)
	}

	err = m.Asset.Validate()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAsset, err.Error())
	}

	// check if amount > 0
	if !m.Amount.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

func (m MsgInboundTransfer) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgOutboundTransfer{}

func NewMsgOutboundTransfer(
	sender string,
	destAddr string,
	asset Asset,
	amount math.Int,
) *MsgOutboundTransfer {
	return &MsgOutboundTransfer{
		Sender:   sender,
		DestAddr: destAddr,
		Asset:    asset,
		Amount:   amount,
	}
}

func (m MsgOutboundTransfer) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(m.DestAddr)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid destination address (%s)", err)
	}

	err = m.Asset.Validate()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAsset, err.Error())
	}

	// check if amount > 0
	if !m.Amount.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

func (m MsgOutboundTransfer) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgUpdateParams{}

func NewMsgUpdateParams(
	sender string,
	newParams Params,
) *MsgUpdateParams {
	return &MsgUpdateParams{
		Sender:    sender,
		NewParams: newParams,
	}
}

func (m MsgUpdateParams) ValidateBasic() error {
	err := m.NewParams.Validate()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidParams, err.Error())
	}
	return nil
}

func (m MsgUpdateParams) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgChangeAssetStatus{}

func NewMsgChangeAssetStatus(
	sender string,
	asset Asset,
	newAssetStatus AssetStatus,
) *MsgChangeAssetStatus {
	return &MsgChangeAssetStatus{
		Sender:         sender,
		Asset:          asset,
		NewAssetStatus: newAssetStatus,
	}
}

func (m MsgChangeAssetStatus) ValidateBasic() error {
	err := m.NewAssetStatus.Validate()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAsset, err.Error())
	}
	return nil
}

func (m MsgChangeAssetStatus) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}
