package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// TODO: 일단 이 메세지 타입들을 사용하되, proto 정의가 완성되면 그것을 사용한다.
var (
	_ sdk.Msg = &MsgSwap{}
	_ sdk.Msg = &MsgJoinPool{}
	_ sdk.Msg = &MsgExitPool{}
	_ sdk.Msg = &MsgCreatePool{}
)

type MsgSwap struct{}

func (m MsgSwap) Reset() {
	panic("implement me")
}

func (m MsgSwap) String() string {
	panic("implement me")
}

func (m MsgSwap) ProtoMessage() {
	panic("implement me")
}

func (m MsgSwap) Route() string {
	panic("implement me")
}

func (m MsgSwap) Type() string {
	panic("implement me")
}

func (m MsgSwap) ValidateBasic() error {
	panic("implement me")
}

func (m MsgSwap) GetSignBytes() []byte {
	panic("implement me")
}

func (m MsgSwap) GetSigners() []sdk.AccAddress {
	panic("implement me")
}

type MsgJoinPool struct{}

func (m MsgJoinPool) Reset() {
	panic("implement me")
}

func (m MsgJoinPool) String() string {
	panic("implement me")
}

func (m MsgJoinPool) ProtoMessage() {
	panic("implement me")
}

func (m MsgJoinPool) Route() string {
	panic("implement me")
}

func (m MsgJoinPool) Type() string {
	panic("implement me")
}

func (m MsgJoinPool) ValidateBasic() error {
	panic("implement me")
}

func (m MsgJoinPool) GetSignBytes() []byte {
	panic("implement me")
}

func (m MsgJoinPool) GetSigners() []sdk.AccAddress {
	panic("implement me")
}

type MsgExitPool struct{}

func (m MsgExitPool) Reset() {
	panic("implement me")
}

func (m MsgExitPool) String() string {
	panic("implement me")
}

func (m MsgExitPool) ProtoMessage() {
	panic("implement me")
}

func (m MsgExitPool) Route() string {
	panic("implement me")
}

func (m MsgExitPool) Type() string {
	panic("implement me")
}

func (m MsgExitPool) ValidateBasic() error {
	panic("implement me")
}

func (m MsgExitPool) GetSignBytes() []byte {
	panic("implement me")
}

func (m MsgExitPool) GetSigners() []sdk.AccAddress {
	panic("implement me")
}

type MsgCreatePool struct{}

func (m MsgCreatePool) Reset() {
	panic("implement me")
}

func (m MsgCreatePool) String() string {
	panic("implement me")
}

func (m MsgCreatePool) ProtoMessage() {
	panic("implement me")
}

func (m MsgCreatePool) Route() string {
	panic("implement me")
}

func (m MsgCreatePool) Type() string {
	panic("implement me")
}

func (m MsgCreatePool) ValidateBasic() error {
	panic("implement me")
}

func (m MsgCreatePool) GetSignBytes() []byte {
	panic("implement me")
}

func (m MsgCreatePool) GetSigners() []sdk.AccAddress {
	panic("implement me")
}
