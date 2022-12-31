package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	// authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*ConcentratedPoolExtension)(nil), nil)
	cdc.RegisterConcrete(&MsgCreatePosition{}, "osmosis/cl-create-position", nil)
	cdc.RegisterConcrete(&MsgWithdrawPosition{}, "osmosis/cl-withdraw-position", nil)
	cdc.RegisterConcrete(&MsgSwapExactAmountIn{}, "osmosis/cl-swap-exact-amount-in", nil)
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
		&MsgSwapExactAmountIn{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
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
