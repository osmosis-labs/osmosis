package model

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	"github.com/osmosis-labs/osmosis/v19/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&CosmWasmPool{}, "osmosis/cw-pool", nil)
	cdc.RegisterConcrete(&Pool{}, "osmosis/cw-pool-wrap", nil)
	cdc.RegisterConcrete(&MsgCreateCosmWasmPool{}, "osmosis/cw-create-pool", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"osmosis.poolmanager.v1beta1.PoolI",
		(*poolmanagertypes.PoolI)(nil),
		&CosmWasmPool{},
	)
	registry.RegisterInterface(
		"osmosis.cosmwasmpool.v1beta1.CosmWasmExtension",
		(*types.CosmWasmExtension)(nil),
		&CosmWasmPool{},
	)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateCosmWasmPool{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_MsgCreator_serviceDesc)
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
