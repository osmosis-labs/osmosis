package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

const (
	setHotRoutes        = "osmosis/MsgSetHotRoutes"
	setDeveloperAccount = "osmosis/MsgSetDeveloperAccount"
)

func init() {
	RegisterCodec(amino)
	amino.Seal()
}

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSetHotRoutes{}, setHotRoutes, nil)
	cdc.RegisterConcrete(&MsgSetDeveloperAccount{}, setDeveloperAccount, nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSetHotRoutes{},
		&MsgSetDeveloperAccount{},
	)
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&SetProtoRevEnabledProposal{},
		&SetProtoRevAdminAccountProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
