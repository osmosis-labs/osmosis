package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgInboundTransfer{}, "osmosis/bridge/inbound-transfer")
	legacy.RegisterAminoMsg(cdc, &MsgOutboundTransfer{}, "osmosis/bridge/outbound-transfer")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "osmosis/bridge/update-params")
	legacy.RegisterAminoMsg(cdc, &MsgChangeAssetStatus{}, "osmosis/bridge/change-asset-status")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgInboundTransfer{},
		&MsgOutboundTransfer{},
		&MsgUpdateParams{},
		&MsgChangeAssetStatus{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)

	// Register all Amino interfaces and concrete types on the authz and gov Amino codec so that this can later be
	// used to properly serialize MsgInboundTransfer, MsgOutboundTransfer,
	// MsgUpdateParams, MsgChangeAssetStatus instances
	RegisterLegacyAminoCodec(authzcodec.Amino)

	amino.Seal()
}
