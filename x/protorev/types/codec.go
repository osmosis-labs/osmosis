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
	// msgs
	setHotRoutes         = "osmosis/MsgSetHotRoutes"
	setDeveloperAccount  = "osmosis/MsgSetDeveloperAccount"
	setMaxRoutesPerTx    = "osmosis/MsgSetMaxRoutesPerTx"
	setMaxRoutesPerBlock = "osmosis/MsgSetMaxRoutesPerBlock"

	// proposals
	setProtoRevEnabledProposal      = "osmosis/SetProtoRevEnabledProposal"
	setProtoRevAdminAccountProposal = "osmosis/SetProtoRevAdminAccountProposal"
)

func init() {
	RegisterCodec(amino)
	amino.Seal()
}

func RegisterCodec(cdc *codec.LegacyAmino) {
	// msgs
	cdc.RegisterConcrete(&MsgSetHotRoutes{}, setHotRoutes, nil)
	cdc.RegisterConcrete(&MsgSetDeveloperAccount{}, setDeveloperAccount, nil)
	cdc.RegisterConcrete(&MsgSetMaxRoutesPerTx{}, setMaxRoutesPerTx, nil)
	cdc.RegisterConcrete(&MsgSetMaxRoutesPerBlock{}, setMaxRoutesPerBlock, nil)

	// proposals
	cdc.RegisterConcrete(&SetProtoRevEnabledProposal{}, setProtoRevEnabledProposal, nil)
	cdc.RegisterConcrete(&SetProtoRevAdminAccountProposal{}, setProtoRevAdminAccountProposal, nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// msgs
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSetHotRoutes{},
		&MsgSetDeveloperAccount{},
		&MsgSetMaxRoutesPerTx{},
		&MsgSetMaxRoutesPerBlock{},
	)

	// proposals
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&SetProtoRevEnabledProposal{},
		&SetProtoRevAdminAccountProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
