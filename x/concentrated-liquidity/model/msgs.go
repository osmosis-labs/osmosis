package model

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	cltypes "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// constants.
const (
	TypeMsgCreateConcentratedPool = "create_concentrated_pool"
	TypeMsgCreatePosition         = "create-position"
	TypeMsgWithdrawPosition       = "withdraw-position"
)

var _ sdk.Msg = &MsgCreatePosition{}

func (msg MsgCreatePosition) Route() string { return cltypes.RouterKey }
func (msg MsgCreatePosition) Type() string  { return TypeMsgCreatePosition }
func (msg MsgCreatePosition) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return fmt.Errorf("Invalid sender address (%s)", err)
	}

	if msg.LowerTick >= msg.UpperTick {
		return cltypes.InvalidLowerUpperTickError{LowerTick: msg.LowerTick, UpperTick: msg.UpperTick}
	}

	if msg.LowerTick < cltypes.MinTick || msg.LowerTick > cltypes.MaxTick {
		return cltypes.InvalidLowerTickError{LowerTick: msg.LowerTick}
	}

	if msg.UpperTick < cltypes.MinTick || msg.UpperTick > cltypes.MaxTick {
		return cltypes.InvalidUpperTickError{UpperTick: msg.UpperTick}
	}

	if !msg.TokenDesired0.IsValid() || msg.TokenDesired0.IsZero() {
		return fmt.Errorf("Invalid coins (%s)", msg.TokenDesired0.String())
	}

	if !msg.TokenDesired1.IsValid() || msg.TokenDesired1.IsZero() {
		return fmt.Errorf("Invalid coins (%s)", msg.TokenDesired1.String())
	}

	if msg.TokenMinAmount0.IsNegative() {
		return cltypes.NotPositiveRequireAmountError{Amount: msg.TokenMinAmount0.String()}
	}

	if msg.TokenMinAmount1.IsNegative() {
		return cltypes.NotPositiveRequireAmountError{Amount: msg.TokenMinAmount1.String()}
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

func (msg MsgWithdrawPosition) Route() string { return cltypes.RouterKey }
func (msg MsgWithdrawPosition) Type() string  { return TypeMsgWithdrawPosition }
func (msg MsgWithdrawPosition) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return fmt.Errorf("Invalid sender address (%s)", err)
	}

	if msg.LowerTick >= msg.UpperTick {
		return cltypes.InvalidLowerUpperTickError{LowerTick: msg.LowerTick, UpperTick: msg.UpperTick}
	}

	if msg.LowerTick < cltypes.MinTick || msg.LowerTick > cltypes.MaxTick {
		return cltypes.InvalidLowerTickError{LowerTick: msg.LowerTick}
	}

	if msg.UpperTick < cltypes.MinTick || msg.UpperTick > cltypes.MaxTick {
		return cltypes.InvalidUpperTickError{UpperTick: msg.UpperTick}
	}

	if !msg.LiquidityAmount.IsPositive() {
		return cltypes.NotPositiveRequireAmountError{Amount: msg.LiquidityAmount.String()}
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

var (
	_ sdk.Msg                       = &MsgCreateConcentratedPool{}
	_ swaproutertypes.CreatePoolMsg = &MsgCreateConcentratedPool{}
)

func NewMsgCreateConcentratedPool(
	sender sdk.AccAddress,
	denom0 string,
	denom1 string,
) MsgCreateConcentratedPool {
	return MsgCreateConcentratedPool{
		Sender: sender.String(),
		Denom0: denom0,
		Denom1: denom1,
	}
}

func (msg MsgCreateConcentratedPool) Route() string { return cltypes.RouterKey }
func (msg MsgCreateConcentratedPool) Type() string  { return TypeMsgCreateConcentratedPool }
func (msg MsgCreateConcentratedPool) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	return nil
}

func (msg MsgCreateConcentratedPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreateConcentratedPool) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

/// Implement the CreatePoolMsg interface

func (msg MsgCreateConcentratedPool) PoolCreator() sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return sender
}

func (msg MsgCreateConcentratedPool) Validate(ctx sdk.Context) error {
	return msg.ValidateBasic()
}

func (msg MsgCreateConcentratedPool) InitialLiquidity() sdk.Coins {
	return nil
}

func (msg MsgCreateConcentratedPool) CreatePool(ctx sdk.Context, poolID uint64) (swaproutertypes.PoolI, error) {
	poolI, err := NewConcentratedLiquidityPool(poolID, msg.Denom0, msg.Denom1)
	return &poolI, err
}

func (msg MsgCreateConcentratedPool) GetPoolType() swaproutertypes.PoolType {
	return swaproutertypes.Concentrated
}
