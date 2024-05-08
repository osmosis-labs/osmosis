package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*CosmWasmExtension)(nil), nil)

	// gov proposals
	cdc.RegisterConcrete(&UploadCosmWasmPoolCodeAndWhiteListProposal{}, "osmosis/upload-cw-pool-code", nil)
	cdc.RegisterConcrete(&MigratePoolContractsProposal{}, "osmosis/migrate-pool-contracts", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface(
		"osmosis.cosmwasmpool.v1beta1.CosmWasmExtension",
		(*CosmWasmExtension)(nil),
	)

	registry.RegisterImplementations(
		(*govtypesv1.Content)(nil),
		&UploadCosmWasmPoolCodeAndWhiteListProposal{},
		&MigratePoolContractsProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
