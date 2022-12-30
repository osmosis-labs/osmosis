package model

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&Pool{}, "osmosis/cl-pool", nil)
	cdc.RegisterConcrete(&MsgCreateConcentratedPool{}, "osmosis/cl-create-pool", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"osmosis.swaprouter.v1beta1.PoolI",
		(*swaproutertypes.PoolI)(nil),
		&Pool{},
	)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateConcentratedPool{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_MsgCreator_serviceDesc)
}

// TODO: re-enable this when CL state-breakage PR is merged.
// return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
// var (
// 	amino     = codec.NewLegacyAmino()
// 	ModuleCdc = codec.NewAminoCodec(amino)
// )

// func init() {
// 	RegisterCodec(amino)
// 	sdk.RegisterLegacyAminoCodec(amino)

// 	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
// 	// used to properly serialize MsgGrant and MsgExec instances
// 	RegisterCodec(authzcodec.Amino)
// 	amino.Seal()
// }
