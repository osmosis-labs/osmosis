package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// constants
const (
	TypeMsgCreatePool              = "create_pool"
	TypeMsgSwapExactAmountIn       = "swap_exact_amount_in"
	TypeMsgSwapExactAmountOut      = "swap_exact_amount_out"
	TypeMsgJoinPool                = "join_pool"
	TypeMsgExitPool                = "exit_pool"
	TypeMsgJoinSwapExternAmountIn  = "join_swap_extern_amount_in"
	TypeMsgJoinSwapShareAmountOut  = "join_swap_share_amount_out"
	TypeMsgExitSwapExternAmountOut = "exit_swap_extern_amount_out"
	TypeMsgExitSwapShareAmountIn   = "exit_swap_share_amount_in"
)

var _ sdk.Msg = &MsgCreatePool{}

func (m MsgCreatePool) Route() string { return RouterKey }
func (m MsgCreatePool) Type() string  { return TypeMsgCreatePool }
func (m MsgCreatePool) ValidateBasic() error {
	return nil // TODO
}
func (m MsgCreatePool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgCreatePool) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgSwapExactAmountIn{}

func (m MsgSwapExactAmountIn) Route() string { return RouterKey }
func (m MsgSwapExactAmountIn) Type() string  { return TypeMsgSwapExactAmountIn }
func (m MsgSwapExactAmountIn) ValidateBasic() error {
	return nil // TODO
}
func (m MsgSwapExactAmountIn) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgSwapExactAmountIn) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgSwapExactAmountOut{}

func (m MsgSwapExactAmountOut) Route() string { return RouterKey }
func (m MsgSwapExactAmountOut) Type() string  { return TypeMsgSwapExactAmountOut }
func (m MsgSwapExactAmountOut) ValidateBasic() error {
	return nil // TODO
}
func (m MsgSwapExactAmountOut) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgSwapExactAmountOut) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgJoinPool{}

func (m MsgJoinPool) Route() string { return RouterKey }
func (m MsgJoinPool) Type() string  { return TypeMsgJoinPool }
func (m MsgJoinPool) ValidateBasic() error {
	return nil // TODO
}
func (m MsgJoinPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgJoinPool) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgExitPool{}

func (m MsgExitPool) Route() string { return RouterKey }
func (m MsgExitPool) Type() string  { return TypeMsgExitPool }
func (m MsgExitPool) ValidateBasic() error {
	return nil // TODO
}
func (m MsgExitPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgExitPool) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgJoinSwapExternAmountIn{}

func (m MsgJoinSwapExternAmountIn) Route() string { return RouterKey }
func (m MsgJoinSwapExternAmountIn) Type() string  { return TypeMsgJoinSwapExternAmountIn }
func (m MsgJoinSwapExternAmountIn) ValidateBasic() error {
	return nil // TODO
}
func (m MsgJoinSwapExternAmountIn) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgJoinSwapExternAmountIn) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgJoinSwapShareAmountOut{}

func (m MsgJoinSwapShareAmountOut) Route() string { return RouterKey }
func (m MsgJoinSwapShareAmountOut) Type() string  { return TypeMsgJoinSwapShareAmountOut }
func (m MsgJoinSwapShareAmountOut) ValidateBasic() error {
	return nil // TODO
}
func (m MsgJoinSwapShareAmountOut) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgJoinSwapShareAmountOut) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgExitSwapExternAmountOut{}

func (m MsgExitSwapExternAmountOut) Route() string { return RouterKey }
func (m MsgExitSwapExternAmountOut) Type() string  { return TypeMsgExitSwapExternAmountOut }
func (m MsgExitSwapExternAmountOut) ValidateBasic() error {
	return nil // TODO
}
func (m MsgExitSwapExternAmountOut) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgExitSwapExternAmountOut) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgExitSwapShareAmountIn{}

func (m MsgExitSwapShareAmountIn) Route() string { return RouterKey }
func (m MsgExitSwapShareAmountIn) Type() string  { return TypeMsgExitSwapShareAmountIn }
func (m MsgExitSwapShareAmountIn) ValidateBasic() error {
	return nil // TODO
}
func (m MsgExitSwapShareAmountIn) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgExitSwapShareAmountIn) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
