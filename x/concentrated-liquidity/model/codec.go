package model

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&Pool{}, "osmosis/concentratedliquidity/ConcentratedLiquidityPool", nil)
	cdc.RegisterConcrete(&MsgCreateConcentratedPool{}, "osmosis/concentratedliquidity/create-concentrated-pool", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
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
