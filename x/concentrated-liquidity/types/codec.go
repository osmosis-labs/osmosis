package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*ConcentratedPoolExtension)(nil), nil)
	cdc.RegisterConcrete(&MsgCreatePosition{}, "osmosis/cl-create-position", nil)
	cdc.RegisterConcrete(&MsgWithdrawPosition{}, "osmosis/cl-withdraw-position", nil)
	cdc.RegisterConcrete(&MsgCollectFees{}, "osmosis/cl-collect-fees", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"osmosis.concentratedliquidity.v1beta1.ConcentratedPoolExtension",
		(*ConcentratedPoolExtension)(nil),
	)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreatePosition{},
		&MsgWithdrawPosition{},
		&MsgCollectFees{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterCodec(amino)
	sdk.RegisterLegacyAminoCodec(amino)

	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	RegisterCodec(authzcodec.Amino)
	amino.Seal()
}
