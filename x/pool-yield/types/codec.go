package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&AddPoolIncentivesProposal{}, "osmosis/AddPoolIncentivesProposal", nil)
	cdc.RegisterConcrete(&RemovePoolIncentivesProposal{}, "osmosis/RemovePoolIncentivesProposal", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&AddPoolIncentivesProposal{},
		&RemovePoolIncentivesProposal{},
	)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/pool-yield module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/pool-yield and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	amino.Seal()
}
