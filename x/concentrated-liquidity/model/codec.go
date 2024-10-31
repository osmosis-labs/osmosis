package model

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&Pool{}, "osmosis/cl-pool", nil)
	cdc.RegisterConcrete(&MsgCreateConcentratedPool{}, "osmosis/cl-create-pool", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"osmosis.swaprouter.v1beta1.PoolI",
		(*poolmanagertypes.PoolI)(nil),
		&Pool{},
	)

	registry.RegisterInterface(
		"osmosis.concentratedliquidity.v1beta1.ConcentratedPoolExtension",
		(*types.ConcentratedPoolExtension)(nil),
		&Pool{},
	)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateConcentratedPool{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
