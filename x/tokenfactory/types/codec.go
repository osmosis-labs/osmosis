package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// this line is used by starport scaffolding # 1
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgCreateDenom{}, "osmosis/tokenfactory/create-denom")
	legacy.RegisterAminoMsg(cdc, &MsgMint{}, "osmosis/tokenfactory/mint")
	legacy.RegisterAminoMsg(cdc, &MsgBurn{}, "osmosis/tokenfactory/burn")
	legacy.RegisterAminoMsg(cdc, &MsgChangeAdmin{}, "osmosis/tokenfactory/change-admin")
	legacy.RegisterAminoMsg(cdc, &MsgSetDenomMetadata{}, "osmosis/tokenfactory/set-denom-metadata")
	legacy.RegisterAminoMsg(cdc, &MsgSetBeforeSendHook{}, "osmosis/tokenfactory/set-bef-send-hook")
	legacy.RegisterAminoMsg(cdc, &MsgForceTransfer{}, "osmosis/tokenfactory/force-transfer")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateDenom{},
		&MsgMint{},
		&MsgBurn{},
		&MsgChangeAdmin{},
		&MsgSetDenomMetadata{},
		&MsgSetBeforeSendHook{},
		&MsgForceTransfer{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
