package types

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/osmomath"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgStakeTokens{}
	_ sdk.Msg = &MsgUnstakeTokens{}
)

const (
	TypeMsgStakeTokens   = "msg_stake_tokens"
	TypeMsgUnstakeTokens = "msg_unstake_tokens"
)

func NewMsgStakeTokens(stakerAddress sdk.AccAddress, amount sdk.Coin) *MsgStakeTokens {
	return &MsgStakeTokens{
		Staker: stakerAddress.String(),
		Amount: amount,
	}
}

// Route Implements Msg
func (msg MsgStakeTokens) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgStakeTokens) Type() string { return TypeMsgStakeTokens }

// GetSignBytes Implements Msg
func (msg MsgStakeTokens) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.NewAminoCodec(amino).MustMarshalJSON(&msg))
}

// GetSigners Implements Msg
func (msg MsgStakeTokens) GetSigners() []sdk.AccAddress {
	staker, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{staker}
}

// ValidateBasic Implements Msg
func (msg MsgStakeTokens) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid staker address (%s)", err)
	}

	if msg.Amount.Amount.LTE(osmomath.ZeroInt()) || msg.Amount.Amount.BigInt().BitLen() > 100 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	return nil
}

func NewMsgUnstakeTokens(stakerAddress sdk.AccAddress, amount sdk.Coin) *MsgUnstakeTokens {
	return &MsgUnstakeTokens{
		Staker: stakerAddress.String(),
		Amount: amount,
	}
}

// Route Implements Msg
func (msg MsgUnstakeTokens) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgUnstakeTokens) Type() string { return TypeMsgUnstakeTokens }

// GetSignBytes Implements Msg
func (msg MsgUnstakeTokens) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.NewAminoCodec(amino).MustMarshalJSON(&msg))
}

// GetSigners Implements Msg
func (msg MsgUnstakeTokens) GetSigners() []sdk.AccAddress {
	staker, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{staker}
}

// ValidateBasic Implements Msg
func (msg MsgUnstakeTokens) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid staker address (%s)", err)
	}

	if msg.Amount.Amount.LTE(osmomath.ZeroInt()) || msg.Amount.Amount.BigInt().BitLen() > 100 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	return nil
}
