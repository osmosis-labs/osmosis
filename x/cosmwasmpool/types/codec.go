package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
		(*govtypes.Content)(nil),
		&UploadCosmWasmPoolCodeAndWhiteListProposal{},
		&MigratePoolContractsProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterCodec(amino)
	RegisterCodec(authzcodec.Amino)

	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	sdk.RegisterLegacyAminoCodec(amino)

	amino.Seal()
}
