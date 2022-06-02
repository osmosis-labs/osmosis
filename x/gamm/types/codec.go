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
	cdc.RegisterConcrete(&Pool{}, "osmosis/gamm/Pool", nil)
	cdc.RegisterConcrete(&MsgCreatePool{}, "osmosis/gamm/create-pool", nil)
	cdc.RegisterConcrete(&MsgJoinPool{}, "osmosis/gamm/join-pool", nil)
	cdc.RegisterConcrete(&MsgExitPool{}, "osmosis/gamm/exit-pool", nil)
	cdc.RegisterConcrete(&MsgSwapExactAmountIn{}, "osmosis/gamm/swap-exact-amount-in", nil)
	cdc.RegisterConcrete(&MsgSwapExactAmountOut{}, "osmosis/gamm/swap-exact-amount-out", nil)
	cdc.RegisterConcrete(&MsgJoinSwapExternAmountIn{}, "osmosis/gamm/join-swap-extern-amount-in", nil)
	cdc.RegisterConcrete(&MsgJoinSwapShareAmountOut{}, "osmosis/gamm/join-swap-share-amount-out", nil)
	cdc.RegisterConcrete(&MsgExitSwapExternAmountOut{}, "osmosis/gamm/exit-swap-extern-amount-out", nil)
	cdc.RegisterConcrete(&MsgExitSwapShareAmountIn{}, "osmosis/gamm/exit-swap-share-amount-in", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface(
		"osmosis.gamm.v1beta1.Pool",
		(*PoolI)(nil),
		&Pool{},
	)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreatePool{},
		&MsgJoinPool{},
		&MsgExitPool{},
		&MsgSwapExactAmountIn{},
		&MsgSwapExactAmountOut{},
		&MsgJoinSwapExternAmountIn{},
		&MsgJoinSwapShareAmountOut{},
		&MsgExitSwapExternAmountOut{},
		&MsgExitSwapShareAmountIn{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/bank module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/staking and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}
