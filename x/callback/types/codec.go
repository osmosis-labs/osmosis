package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptoCodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types on the provided LegacyAmino codec.
// These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRequestCallback{}, "callback/MsgRequestCallback", nil)
	cdc.RegisterConcrete(&MsgCancelCallback{}, "callback/MsgCancelCallback", nil)
}

// RegisterInterfaces registers interfaces types with the interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRequestCallback{},
		&MsgCancelCallback{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
	amino     = codec.NewLegacyAmino()
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptoCodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)
	amino.Seal()
}
