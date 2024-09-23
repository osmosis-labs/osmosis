package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/osmomath"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgSwap{}
	_ sdk.Msg = &MsgSwapSend{}
)

// market message types
const (
	TypeMsgSwap     = "swap"
	TypeMsgSwapSend = "swap_send"
)

//--------------------------------------------------------
//--------------------------------------------------------

// NewMsgSwap creates a MsgSwap instance
func NewMsgSwap(traderAddress sdk.AccAddress, offerCoin sdk.Coin, askCoin string) *MsgSwap {
	return &MsgSwap{
		Trader:    traderAddress.String(),
		OfferCoin: offerCoin,
		AskDenom:  askCoin,
	}
}

// Route Implements Msg
func (msg MsgSwap) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgSwap) Type() string { return TypeMsgSwap }

// GetSignBytes Implements Msg
func (msg MsgSwap) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners Implements Msg
func (msg MsgSwap) GetSigners() []sdk.AccAddress {
	trader, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{trader}
}

// ValidateBasic Implements Msg
func (msg MsgSwap) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid trader address (%s)", err)
	}

	if msg.OfferCoin.Amount.LTE(osmomath.ZeroInt()) || msg.OfferCoin.Amount.BigInt().BitLen() > 100 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, msg.OfferCoin.String())
	}

	if msg.OfferCoin.Denom == msg.AskDenom {
		return errorsmod.Wrap(ErrRecursiveSwap, msg.AskDenom)
	}

	return nil
}

// NewMsgSwapSend conducts market swap and send all the result coins to recipient
func NewMsgSwapSend(fromAddress sdk.AccAddress, toAddress sdk.AccAddress, offerCoin sdk.Coin, askCoin string) *MsgSwapSend {
	return &MsgSwapSend{
		FromAddress: fromAddress.String(),
		ToAddress:   toAddress.String(),
		OfferCoin:   offerCoin,
		AskDenom:    askCoin,
	}
}

// Route Implements Msg
func (msg MsgSwapSend) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgSwapSend) Type() string { return TypeMsgSwapSend }

// GetSignBytes Implements Msg
func (msg MsgSwapSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners Implements Msg
func (msg MsgSwapSend) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{from}
}

// ValidateBasic Implements Msg
func (msg MsgSwapSend) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid from address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid to address (%s)", err)
	}

	if msg.OfferCoin.Amount.LTE(osmomath.ZeroInt()) || msg.OfferCoin.Amount.BigInt().BitLen() > 100 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, msg.OfferCoin.String())
	}

	if msg.OfferCoin.Denom == msg.AskDenom {
		return errorsmod.Wrap(ErrRecursiveSwap, msg.AskDenom)
	}

	return nil
}
