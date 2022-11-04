package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// constants.
const (
	TypeMsgCreatePosition   = "create-position"
	TypeMsgWithdrawPosition = "withdraw-position"
)

var _ sdk.Msg = &MsgCreatePosition{}

func (msg MsgCreatePosition) Route() string { return RouterKey }
func (msg MsgCreatePosition) Type() string  { return TypeMsgCreatePosition }
func (msg MsgCreatePosition) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if msg.LowerTick >= msg.UpperTick {
		return InvalidLowerUpperTickError{LowerTick: msg.LowerTick, UpperTick: msg.UpperTick}
	}

	if msg.LowerTick < MinTick || msg.LowerTick > MaxTick {
		return InvalidLowerTickError{LowerTick: msg.LowerTick}
	}

	if msg.UpperTick < MinTick || msg.UpperTick > MaxTick {
		return InvalidUpperTickError{UpperTick: msg.UpperTick}
	}

	if !msg.TokenDesired0.IsValid() || msg.TokenDesired0.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.TokenDesired0.String())
	}

	if !msg.TokenDesired1.IsValid() || msg.TokenDesired1.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.TokenDesired1.String())
	}

	if msg.TokenMinAmount0.IsNegative() {
		return NotPositiveRequireAmountError{Amount: msg.TokenMinAmount0.String()}
	}

	if msg.TokenMinAmount1.IsNegative() {
		return NotPositiveRequireAmountError{Amount: msg.TokenMinAmount1.String()}
	}

	return nil
}

func (msg MsgCreatePosition) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreatePosition) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgWithdrawPosition{}

func (msg MsgWithdrawPosition) Route() string { return RouterKey }
func (msg MsgWithdrawPosition) Type() string  { return TypeMsgWithdrawPosition }
func (msg MsgWithdrawPosition) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if msg.LowerTick >= msg.UpperTick {
		return InvalidLowerUpperTickError{LowerTick: msg.LowerTick, UpperTick: msg.UpperTick}
	}

	if msg.LowerTick < MinTick || msg.LowerTick > MaxTick {
		return InvalidLowerTickError{LowerTick: msg.LowerTick}
	}

	if msg.UpperTick < MinTick || msg.UpperTick > MaxTick {
		return InvalidUpperTickError{UpperTick: msg.UpperTick}
	}

	if !msg.LiquidityAmount.IsPositive() {
		return NotPositiveRequireAmountError{Amount: msg.LiquidityAmount.String()}
	}

	return nil
}

func (msg MsgWithdrawPosition) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgWithdrawPosition) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
