package model

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	cltypes "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// constants.
const (
	TypeMsgCreateConcentratedPool = "create_concentrated_pool"
)

var (
	_ sdk.Msg                       = &MsgCreateConcentratedPool{}
	_ swaproutertypes.CreatePoolMsg = &MsgCreateConcentratedPool{}
)

func NewMsgCreateConcentratedPool(
	sender sdk.AccAddress,
	denom0 string,
	denom1 string,
	tickSpacing uint64,
) MsgCreateConcentratedPool {
	return MsgCreateConcentratedPool{
		Sender:      sender.String(),
		Denom0:      denom0,
		Denom1:      denom1,
		TickSpacing: tickSpacing,
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
	return sdk.Coins{}
}

func (msg MsgCreateConcentratedPool) CreatePool(ctx sdk.Context, poolID uint64) (swaproutertypes.PoolI, error) {
	poolI, err := NewConcentratedLiquidityPool(poolID, msg.Denom0, msg.Denom1, msg.TickSpacing)
	return &poolI, err
}

func (msg MsgCreateConcentratedPool) GetPoolType() swaproutertypes.PoolType {
	return swaproutertypes.Concentrated
}
