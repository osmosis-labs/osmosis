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

// TODO: 일단 이 메세지 타입들을 사용하되, proto 정의가 완성되면 그것을 사용한다.
var (
	_ sdk.Msg = &MsgCreatePool{}
	_ sdk.Msg = &MsgSwapExactAmountIn{}
	_ sdk.Msg = &MsgSwapExactAmountOut{}
	_ sdk.Msg = &MsgJoinPool{}
	_ sdk.Msg = &MsgExitPool{}
)

type MsgCreatePool struct {
	Sender  sdk.AccAddress `json:"sender"`
	SwapFee sdk.Dec        `json:"swap_fee"`

	TokenInfo struct {
		Token  sdk.Coin `json:"token"`
		Ratio  sdk.Dec  `json:"ratio"`
		Amount sdk.Int  `json:"amount"`
	} `json:"token_info"`
}

func (m MsgCreatePool) Reset()         { panic("implement me") }
func (m MsgCreatePool) String() string { panic("implement me") }
func (m MsgCreatePool) ProtoMessage()  { panic("implement me") }

func (m MsgCreatePool) Route() string { return RouterKey }
func (m MsgCreatePool) Type() string  { return TypeMsgCreatePool }
func (m MsgCreatePool) ValidateBasic() error {
	return nil // TODO
}
func (m MsgCreatePool) GetSignBytes() []byte {
	return nil // TODO
}
func (m MsgCreatePool) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }

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

type MsgJoinPool struct {
	Sender        sdk.AccAddress `json:"sender"`
	TargetPool    sdk.AccAddress `json:"target_pool"`
	PoolAmountOut sdk.Int        `json:"pool_amount_out"`
	MaxAmountsIn  struct {
		Token     string  `json:"token"`
		MaxAmount sdk.Int `json:"max_amount"`
	} `json:"max_amounts_in"`
}

func (m MsgJoinPool) Reset()         { panic("implement me") }
func (m MsgJoinPool) String() string { panic("implement me") }
func (m MsgJoinPool) ProtoMessage()  { panic("implement me") }

func (m MsgJoinPool) Route() string { return RouterKey }
func (m MsgJoinPool) Type() string  { return TypeMsgJoinPool }
func (m MsgJoinPool) ValidateBasic() error {
	return nil // TODO
}
func (m MsgJoinPool) GetSignBytes() []byte {
	return nil // TODO
}
func (m MsgJoinPool) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }

type MsgExitPool struct {
	Sender        sdk.AccAddress `json:"sender"`
	TargetPool    sdk.AccAddress `json:"target_pool"`
	PoolAmountIn  sdk.Int        `json:"pool_amount_out"`
	MinAmountsOut struct {
		Token     string  `json:"token"`
		MinAmount sdk.Int `json:"min_amount"`
	} `json:"min_amounts_out"`
}

func (m MsgExitPool) Reset()         { panic("implement me") }
func (m MsgExitPool) String() string { panic("implement me") }
func (m MsgExitPool) ProtoMessage()  { panic("implement me") }

func (m MsgExitPool) Route() string { return RouterKey }
func (m MsgExitPool) Type() string  { return TypeMsgExitPool }
func (m MsgExitPool) ValidateBasic() error {
	return nil // TODO
}
func (m MsgExitPool) GetSignBytes() []byte {
	return nil // TODO
}
func (m MsgExitPool) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
