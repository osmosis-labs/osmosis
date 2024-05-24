package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// RegisterCodec registers the necessary x/incentives interfaces and concrete types on the provided
// LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateGauge{}, "osmosis/incentives/create-gauge", nil)
	cdc.RegisterConcrete(&MsgAddToGauge{}, "osmosis/incentives/add-to-gauge", nil)

	// gov proposals
	cdc.RegisterConcrete(&CreateGroupsProposal{}, "osmosis/create-groups-proposal", nil)
}

// RegisterInterfaces registers interfaces and implementations of the incentives module.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateGauge{},
		&MsgAddToGauge{},
	)

	registry.RegisterImplementations(
		(*govtypesv1.Content)(nil),
		&CreateGroupsProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
