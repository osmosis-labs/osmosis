package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

const (
	// msgs
	setHotRoutes             = "osmosis/MsgSetHotRoutes"
	setDeveloperAccount      = "osmosis/MsgSetDeveloperAccount"
	setMaxPoolPointsPerTx    = "osmosis/MsgSetMaxPoolPointsPerTx"
	setMaxPoolPointsPerBlock = "osmosis/MsgSetMaxPoolPointsPerBlock"
	setInfoByPoolType        = "osmosis/MsgSetInfoByPoolType"
	setBaseDenoms            = "osmosis/MsgSetBaseDenoms"

	// proposals
	setProtoRevEnabledProposal      = "osmosis/SetProtoRevEnabledProposal"
	setProtoRevAdminAccountProposal = "osmosis/SetProtoRevAdminAccountProposal"
)

func init() {
	RegisterCodec(amino)
	sdk.RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

func RegisterCodec(cdc *codec.LegacyAmino) {
	// msgs
	cdc.RegisterConcrete(&MsgSetHotRoutes{}, setHotRoutes, nil)
	cdc.RegisterConcrete(&MsgSetDeveloperAccount{}, setDeveloperAccount, nil)
	cdc.RegisterConcrete(&MsgSetMaxPoolPointsPerTx{}, setMaxPoolPointsPerTx, nil)
	cdc.RegisterConcrete(&MsgSetMaxPoolPointsPerBlock{}, setMaxPoolPointsPerBlock, nil)
	cdc.RegisterConcrete(&MsgSetInfoByPoolType{}, setInfoByPoolType, nil)
	cdc.RegisterConcrete(&MsgSetBaseDenoms{}, setBaseDenoms, nil)

	// proposals
	cdc.RegisterConcrete(&SetProtoRevEnabledProposal{}, setProtoRevEnabledProposal, nil)
	cdc.RegisterConcrete(&SetProtoRevAdminAccountProposal{}, setProtoRevAdminAccountProposal, nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// msgs
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSetHotRoutes{},
		&MsgSetDeveloperAccount{},
		&MsgSetMaxPoolPointsPerTx{},
		&MsgSetMaxPoolPointsPerBlock{},
		&MsgSetInfoByPoolType{},
		&MsgSetBaseDenoms{},
	)

	// proposals
	registry.RegisterImplementations(
		(*govtypesv1.Content)(nil),
		&SetProtoRevEnabledProposal{},
		&SetProtoRevAdminAccountProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
