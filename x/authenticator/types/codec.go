package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

type AuthenticatorTxOptions interface {
	GetSelectedAuthenticators() []int32
}

func RegisterCodec(cdc *codec.LegacyAmino) {
	//cdc.RegisterConcrete(&TxExtension{}, "osmosis.authenticator.TxExtension", nil)

}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*AuthenticatorTxOptions)(nil),
		&TxExtension{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
