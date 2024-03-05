package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgInboundTransfer{}, "osmosis/bridge/inbound-transfer")
	legacy.RegisterAminoMsg(cdc, &MsgOutboundTransfer{}, "osmosis/bridge/outbound-transfer")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "osmosis/bridge/update-params")
	legacy.RegisterAminoMsg(cdc, &MsgEnableBridge{}, "osmosis/bridge/enable-bridge")
	legacy.RegisterAminoMsg(cdc, &MsgDisableBridge{}, "osmosis/bridge/disable-bridge")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgInboundTransfer{},
		&MsgOutboundTransfer{},
		&MsgUpdateParams{},
		&MsgEnableBridge{},
		&MsgDisableBridge{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	// TODO: complete
	RegisterLegacyAminoCodec(amino)
	sdk.RegisterLegacyAminoCodec(amino)
	amino.Seal()
}
