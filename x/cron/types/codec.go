package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterCron{}, "osmosis/cron/MsgRegisterCron", nil)
	cdc.RegisterConcrete(&MsgUpdateCronJob{}, "osmosis/cron/MsgUpdateCronJob", nil)
	cdc.RegisterConcrete(&MsgDeleteCronJob{}, "osmosis/cron/MsgDeleteCronJob", nil)
	cdc.RegisterConcrete(&MsgToggleCronJob{}, "osmosis/cron/MsgToggleCronJob", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterCron{},
		&MsgUpdateCronJob{},
		&MsgDeleteCronJob{},
		&MsgToggleCronJob{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(Amino)
	cryptocodec.RegisterCrypto(Amino)
	Amino.Seal()
}
