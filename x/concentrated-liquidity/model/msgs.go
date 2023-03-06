package model

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// constants.
const (
	TypeMsgCreateConcentratedPool = "create_concentrated_pool"
)

var (
	_ sdk.Msg                        = &MsgCreateConcentratedPool{}
	_ poolmanagertypes.CreatePoolMsg = &MsgCreateConcentratedPool{}
)

func NewMsgCreateConcentratedPool(
	sender sdk.AccAddress,
	denom0 string,
	denom1 string,
	tickSpacing uint64,
	precisionFactorAtPriceOne sdk.Int,
	swapFee sdk.Dec,
) MsgCreateConcentratedPool {
	return MsgCreateConcentratedPool{
		Sender:                    sender.String(),
		Denom0:                    denom0,
		Denom1:                    denom1,
		TickSpacing:               tickSpacing,
		PrecisionFactorAtPriceOne: precisionFactorAtPriceOne,
		SwapFee:                   swapFee,
	}
}

func (msg MsgCreateConcentratedPool) Route() string { return cltypes.RouterKey }
func (msg MsgCreateConcentratedPool) Type() string  { return TypeMsgCreateConcentratedPool }
func (msg MsgCreateConcentratedPool) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if msg.TickSpacing <= 0 {
		return fmt.Errorf("tick spacing must be positive")
	}

	if msg.Denom0 == msg.Denom1 {
		return fmt.Errorf("denom0 and denom1 must be different")
	}

	if msg.PrecisionFactorAtPriceOne.GT(cltypes.ExponentAtPriceOneMax) || msg.PrecisionFactorAtPriceOne.LT(cltypes.ExponentAtPriceOneMin) {
		return cltypes.ExponentAtPriceOneError{ProvidedExponentAtPriceOne: msg.PrecisionFactorAtPriceOne, PrecisionValueAtPriceOneMin: cltypes.ExponentAtPriceOneMin, PrecisionValueAtPriceOneMax: cltypes.ExponentAtPriceOneMax}
	}

	if sdk.ValidateDenom(msg.Denom0) != nil {
		return fmt.Errorf("denom0 is invalid: %s", sdk.ValidateDenom(msg.Denom0))
	}

	if sdk.ValidateDenom(msg.Denom1) != nil {
		return fmt.Errorf("denom1 is invalid: %s", sdk.ValidateDenom(msg.Denom1))
	}

	swapFee := msg.SwapFee
	if swapFee.IsNegative() || swapFee.GTE(one) {
		return cltypes.InvalidSwapFeeError{ActualFee: swapFee}
	}

	return nil
}

func (msg MsgCreateConcentratedPool) GetSignBytes() []byte {
	return nil
	// return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
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

func (msg MsgCreateConcentratedPool) CreatePool(ctx sdk.Context, poolID uint64) (poolmanagertypes.PoolI, error) {
	poolI, err := NewConcentratedLiquidityPool(poolID, msg.Denom0, msg.Denom1, msg.TickSpacing, msg.PrecisionFactorAtPriceOne, msg.SwapFee)
	return &poolI, err
}

func (msg MsgCreateConcentratedPool) GetPoolType() poolmanagertypes.PoolType {
	return poolmanagertypes.Concentrated
}
