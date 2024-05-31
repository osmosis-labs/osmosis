package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// RegisterLegacyAminoCodec registers the necessary x/gamm interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*CFMMPoolI)(nil), nil)
	cdc.RegisterConcrete(&MsgJoinPool{}, "symphony/gamm/join-pool", nil)
	cdc.RegisterConcrete(&MsgExitPool{}, "symphony/gamm/exit-pool", nil)
	cdc.RegisterConcrete(&MsgSwapExactAmountIn{}, "symphony/gamm/swap-exact-amount-in", nil)
	cdc.RegisterConcrete(&MsgSwapExactAmountOut{}, "symphony/gamm/swap-exact-amount-out", nil)
	cdc.RegisterConcrete(&MsgJoinSwapExternAmountIn{}, "symphony/gamm/join-swap-extern-amount-in", nil)
	cdc.RegisterConcrete(&MsgJoinSwapShareAmountOut{}, "symphony/gamm/join-swap-share-amount-out", nil)
	cdc.RegisterConcrete(&MsgExitSwapExternAmountOut{}, "symphony/gamm/exit-swap-extern-amount-out", nil)
	cdc.RegisterConcrete(&MsgExitSwapShareAmountIn{}, "symphony/gamm/exit-swap-share-amount-in", nil)
	cdc.RegisterConcrete(&UpdateMigrationRecordsProposal{}, "symphony/gamm/update-migration-records-proposal", nil)
	cdc.RegisterConcrete(&ReplaceMigrationRecordsProposal{}, "symphony/gamm/replace-migration-records-proposal", nil)
	cdc.RegisterConcrete(&CreateConcentratedLiquidityPoolsAndLinktoCFMMProposal{}, "symphony/gamm/create-cl-pool-and-cfmm-link", nil)
	cdc.RegisterConcrete(&SetScalingFactorControllerProposal{}, "symphony/gamm/scaling-factor-controller", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface(
		"osmosis.gamm.v1beta1.PoolI", // N.B.: the old proto-path is preserved for backwards-compatibility.
		(*CFMMPoolI)(nil),
	)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgJoinPool{},
		&MsgExitPool{},
		&MsgSwapExactAmountIn{},
		&MsgSwapExactAmountOut{},
		&MsgJoinSwapExternAmountIn{},
		&MsgJoinSwapShareAmountOut{},
		&MsgExitSwapExternAmountOut{},
		&MsgExitSwapShareAmountIn{},
	)

	registry.RegisterImplementations(
		(*govtypesv1.Content)(nil),
		&UpdateMigrationRecordsProposal{},
		&ReplaceMigrationRecordsProposal{},
		&CreateConcentratedLiquidityPoolsAndLinktoCFMMProposal{},
		&SetScalingFactorControllerProposal{},
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
	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	sdk.RegisterLegacyAminoCodec(amino)
	RegisterLegacyAminoCodec(authzcodec.Amino)

	amino.Seal()
}
