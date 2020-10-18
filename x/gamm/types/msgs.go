package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgCreatePool         = "create_pool"
	TypeMsgSwapExactAmountIn  = "swap_exact_amount_in"
	TypeMsgSwapExactAmountOut = "swap_exact_amount_out"
	TypeMsgJoinPool           = "join_pool"
	TypeMsgExitPool           = "exit_pool"
)

var _ sdk.Msg = &MsgCreatePool{}

func (m MsgCreatePool) Route() string { return RouterKey }
func (m MsgCreatePool) Type() string  { return TypeMsgCreatePool }
func (m MsgCreatePool) ValidateBasic() error {
	return nil // TODO
}
func (m MsgCreatePool) GetSignBytes() []byte {
	return nil // TODO
}
func (m MsgCreatePool) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }

var _ sdk.Msg = &MsgSwapExactAmountIn{}

type MsgSwapExactAmountIn struct {
	Sender        sdk.AccAddress `json:"sender"`
	TargetPool    sdk.AccAddress `json:"target_pool"`
	TokenIn       sdk.Coin       `json:"token_in"`
	TokenAmountIn sdk.Int        `json:"token_amount_in"`
	TokenOut      sdk.Coin       `json:"token_out"`
	MinAmountOut  sdk.Int        `json:"min_amount_out"`
	MaxPrice      sdk.Int        `json:"max_price"`
}

func (m MsgSwapExactAmountIn) Reset()         { panic("implement me") }
func (m MsgSwapExactAmountIn) String() string { panic("implement me") }
func (m MsgSwapExactAmountIn) ProtoMessage()  { panic("implement me") }

func (m MsgSwapExactAmountIn) Route() string { return RouterKey }
func (m MsgSwapExactAmountIn) Type() string  { return TypeMsgSwapExactAmountIn }
func (m MsgSwapExactAmountIn) ValidateBasic() error {
	return nil // TODO
}
func (m MsgSwapExactAmountIn) GetSignBytes() []byte {
	return nil // TODO
}
func (m MsgSwapExactAmountIn) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }

var _ sdk.Msg = &MsgSwapExactAmountOut{}

type MsgSwapExactAmountOut struct {
	Sender         sdk.AccAddress `json:"sender"`
	TargetPool     sdk.AccAddress `json:"target_pool"`
	TokenIn        sdk.Coin       `json:"token_in"`
	MaxAmountIn    sdk.Int        `json:"max_amount_in"`
	TokenOut       sdk.Coin       `json:"token_out"`
	TokenAmountOut sdk.Int        `json:"token_amount_out"`
	MaxPrice       sdk.Int        `json:"max_price"`
}

func (m MsgSwapExactAmountOut) Reset()         { panic("implement me") }
func (m MsgSwapExactAmountOut) String() string { panic("implement me") }
func (m MsgSwapExactAmountOut) ProtoMessage()  { panic("implement me") }

func (m MsgSwapExactAmountOut) Route() string { return RouterKey }
func (m MsgSwapExactAmountOut) Type() string  { return TypeMsgSwapExactAmountOut }
func (m MsgSwapExactAmountOut) ValidateBasic() error {
	panic("implement me")
}
func (m MsgSwapExactAmountOut) GetSignBytes() []byte {
	panic("implement me")
}
func (m MsgSwapExactAmountOut) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }

var _ sdk.Msg = &MsgJoinPool{}

func (m MsgJoinPool) Route() string { return RouterKey }
func (m MsgJoinPool) Type() string  { return TypeMsgJoinPool }
func (m MsgJoinPool) ValidateBasic() error {
	return nil // TODO
}
func (m MsgJoinPool) GetSignBytes() []byte {
	return nil // TODO
}
func (m MsgJoinPool) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }

var _ sdk.Msg = &MsgExitPool{}

func (m MsgExitPool) Route() string { return RouterKey }
func (m MsgExitPool) Type() string  { return TypeMsgExitPool }
func (m MsgExitPool) ValidateBasic() error {
	return nil // TODO
}
func (m MsgExitPool) GetSignBytes() []byte {
	return nil // TODO
}
func (m MsgExitPool) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
