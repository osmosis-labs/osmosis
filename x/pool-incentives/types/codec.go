package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&AddPoolIncentivesProposal{}, "osmosis/AddPoolIncentivesProposal", nil)
	cdc.RegisterConcrete(&EditPoolIncentivesProposal{}, "osmosis/EditPoolIncentivesProposal", nil)
	cdc.RegisterConcrete(&RemovePoolIncentivesProposal{}, "osmosis/RemovePoolIncentivesProposal", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&AddPoolIncentivesProposal{},
		&EditPoolIncentivesProposal{},
		&RemovePoolIncentivesProposal{},
	)
}
