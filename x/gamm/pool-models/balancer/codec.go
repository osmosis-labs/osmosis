package balancer

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
<<<<<<< HEAD
	types "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
=======
	proto "github.com/gogo/protobuf/proto"
>>>>>>> b2ae53b (fix: pool params query (#1315))
)

// RegisterLegacyAminoCodec registers the necessary x/gamm interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&Pool{}, "osmosis/gamm/BalancerPool", nil)
	cdc.RegisterConcrete(&MsgCreateBalancerPool{}, "osmosis/gamm/create-balancer-pool", nil)
	cdc.RegisterConcrete(&PoolParams{}, "osmosis/gamm/BalancerPoolParams", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"osmosis.gamm.v1beta1.PoolI",
		(*types.PoolI)(nil),
		&Pool{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateBalancerPool{},
	)
	registry.RegisterImplementations(
		(*proto.Message)(nil),
		&PoolParams{},
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
