package model

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*types.ConcentratedPoolExtension)(nil), nil)
	cdc.RegisterConcrete(&MsgCreatePosition{}, "osmosis/cl-create-position", nil)
	cdc.RegisterConcrete(&MsgWithdrawPosition{}, "osmosis/cl-withdraw-position", nil)
	cdc.RegisterConcrete(&MsgCreateConcentratedPool{}, "osmosis/create-concentrated-pool", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreatePosition{},
		&MsgWithdrawPosition{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateConcentratedPool{},
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
