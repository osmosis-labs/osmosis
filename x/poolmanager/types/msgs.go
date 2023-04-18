package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// constants.
const (
	TypeMsgSwapExactAmountIn            = "swap_exact_amount_in"
	TypeMsgSwapExactAmountOut           = "swap_exact_amount_out"
	TypeMsgSplitRouteSwapExactAmountIn  = "split_route_swap_exact_amount_in"
	TypeMsgSplitRouteSwapExactAmountOut = "split_route_swap_exact_amount_out"
)

var _ sdk.Msg = &MsgSwapExactAmountIn{}

func (msg MsgSwapExactAmountIn) Route() string { return RouterKey }
func (msg MsgSwapExactAmountIn) Type() string  { return TypeMsgSwapExactAmountIn }
func (msg MsgSwapExactAmountIn) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	err = SwapAmountInRoutes(msg.Routes).Validate()
	if err != nil {
		return err
	}

	if !msg.TokenIn.IsValid() || !msg.TokenIn.IsPositive() {
		// TODO: remove sdk errors
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.TokenIn.String())
	}

	if !msg.TokenOutMinAmount.IsPositive() {
		return nonPositiveAmountError{msg.TokenOutMinAmount.String()}
	}

	return nil
}

func (msg MsgSwapExactAmountIn) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgSwapExactAmountIn) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgSwapExactAmountOut{}

func (msg MsgSwapExactAmountOut) Route() string { return RouterKey }
func (msg MsgSwapExactAmountOut) Type() string  { return TypeMsgSwapExactAmountOut }
func (msg MsgSwapExactAmountOut) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	err = SwapAmountOutRoutes(msg.Routes).Validate()
	if err != nil {
		return err
	}

	if !msg.TokenOut.IsValid() || !msg.TokenOut.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.TokenOut.String())
	}

	if !msg.TokenInMaxAmount.IsPositive() {
		return nonPositiveAmountError{msg.TokenInMaxAmount.String()}
	}

	return nil
}

func (msg MsgSwapExactAmountOut) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgSwapExactAmountOut) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgSplitRouteSwapExactAmountIn{}

func (msg MsgSplitRouteSwapExactAmountIn) Route() string { return RouterKey }
func (msg MsgSplitRouteSwapExactAmountIn) Type() string  { return TypeMsgSplitRouteSwapExactAmountIn }

func (msg MsgSplitRouteSwapExactAmountIn) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return InvalidSenderError{Sender: msg.Sender}
	}

	if err := sdk.ValidateDenom(msg.TokenInDenom); err != nil {
		return err
	}

	if err := ValidateSwapAmountInSplitRoute(msg.Routes); err != nil {
		return err
	}

	if !msg.TokenOutMinAmount.IsPositive() {
		return nonPositiveAmountError{msg.TokenOutMinAmount.String()}
	}

	return nil
}

func (msg MsgSplitRouteSwapExactAmountIn) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgSplitRouteSwapExactAmountIn) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgSplitRouteSwapExactAmountOut{}

func (msg MsgSplitRouteSwapExactAmountOut) Route() string { return RouterKey }
func (msg MsgSplitRouteSwapExactAmountOut) Type() string  { return TypeMsgSplitRouteSwapExactAmountOut }

func (msg MsgSplitRouteSwapExactAmountOut) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return InvalidSenderError{Sender: msg.Sender}
	}

	if err := sdk.ValidateDenom(msg.TokenOutDenom); err != nil {
		return err
	}

	if err := ValidateSwapAmountOutSplitRoute(msg.Routes); err != nil {
		return err
	}

	if !msg.TokenInMaxAmount.IsPositive() {
		return nonPositiveAmountError{msg.TokenInMaxAmount.String()}
	}

	return nil
}

func (msg MsgSplitRouteSwapExactAmountOut) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgSplitRouteSwapExactAmountOut) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
