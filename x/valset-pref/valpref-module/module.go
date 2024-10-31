package validator_preference

import (
	"context"
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"cosmossdk.io/core/appmodule"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v27/simulation/simtypes"
	keeper "github.com/osmosis-labs/osmosis/v27/x/valset-pref"
	validatorprefclient "github.com/osmosis-labs/osmosis/v27/x/valset-pref/client"
	valsetprefcli "github.com/osmosis-labs/osmosis/v27/x/valset-pref/client/cli"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/client/grpc"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"
)

var (
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ module.HasGenesisBasics = AppModuleBasic{}

	_ appmodule.AppModule        = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic implements the AppModuleBasic interface for the capability module.
type AppModuleBasic struct {
	cdc codec.Codec
}

func NewAppModuleBasic(cdc codec.Codec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

// Name returns the capability module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) RegisterCodec(cdc *codec.LegacyAmino) {
	types.RegisterCodec(cdc)
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterCodec(cdc)
}

// RegisterInterfaces registers the module's interface types.
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis returns the capability module's default genesis state.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return nil
}

// ValidateGenesis performs genesis state validation for the capability module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	queryproto.RegisterQueryHandlerClient(context.Background(), mux, queryproto.NewQueryClient(clientCtx)) //nolint:errcheck
}

// GetTxCmd returns the capability module's root tx command.
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return valsetprefcli.GetTxCmd()
}

// GetQueryCmd returns the capability module's root query command.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return valsetprefcli.GetQueryCmd()
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface for the capability module.
type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// IsOnePerModuleType is a marker function just indicates that this is a one-per-module type.
func (am AppModule) IsOnePerModuleType() {}

// Name returns the capability module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// QuerierRoute returns the capability module's query routing key.
func (AppModule) QuerierRoute() string { return types.QuerierRoute }

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(&am.keeper))
	queryproto.RegisterQueryServer(cfg.QueryServer(), grpc.Querier{Q: validatorprefclient.Querier{K: am.keeper}})
}

// RegisterInvariants registers the capability module's invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
}

// InitGenesis performs the capability module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
}

// ExportGenesis returns the capability module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return nil
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// ___________________________________________________________________________

// AppModuleSimulation functions

// WeightedOperations returns the all the valset module operations with their respective weights.
func (am AppModule) Actions() []simtypes.Action {
	return []simtypes.Action{
		// simtypes.NewMsgBasedAction("SetValidatorSetPreference", am.keeper, simulation.RandomMsgSetValSetPreference),
		// simtypes.NewMsgBasedAction("MsgDelegateToValidatorSet", am.keeper, simulation.RandomMsgDelegateToValSet),
		// simtypes.NewMsgBasedAction("MsgUndelegateFromValidatorSet", am.keeper, simulation.RandomMsgUnDelegateFromValSet),
		// simtypes.NewMsgBasedAction("MsgRedelegateValSet", am.keeper, simulation.RandomMsgReDelegateToValSet),
	}
}
