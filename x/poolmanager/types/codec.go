package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/gamm interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSwapExactAmountIn{}, "osmosis/poolmanager/swap-exact-amount-in", nil)
	cdc.RegisterConcrete(&MsgSwapExactAmountOut{}, "osmosis/poolmanager/swap-exact-amount-out", nil)
	cdc.RegisterConcrete(&MsgSplitRouteSwapExactAmountIn{}, "osmosis/poolmanager/split-amount-in", nil)
	cdc.RegisterConcrete(&MsgSplitRouteSwapExactAmountOut{}, "osmosis/poolmanager/split-amount-out", nil)
	cdc.RegisterConcrete(&MsgSetRevenueShareUser{}, "osmosis/poolmanager/revenue-share", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSwapExactAmountIn{},
		&MsgSwapExactAmountOut{},
		&MsgSplitRouteSwapExactAmountIn{},
		&MsgSplitRouteSwapExactAmountOut{},
		&MsgSetRevenueShareUser{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
