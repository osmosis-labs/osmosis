package module

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"

	"github.com/osmosis-labs/osmosis/v27/simulation/simtypes"
	gammsimulation "github.com/osmosis-labs/osmosis/v27/x/gamm/simulation"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager"
	pmclient "github.com/osmosis-labs/osmosis/v27/x/poolmanager/client"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/cli"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/grpc"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/grpcv2"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/queryprotov2"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

var (
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ module.HasGenesisBasics = AppModuleBasic{}

	_ appmodule.AppModule        = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}
)

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the poolmanager module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// ---------------------------------------
// Interfaces.
func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := queryproto.RegisterQueryHandlerClient(context.Background(), mux, queryproto.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
	if err := queryprotov2.RegisterQueryHandlerClient(context.Background(), mux, queryprotov2.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// RegisterInterfaces registers interfaces and implementations of the gamm module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

type AppModule struct {
	AppModuleBasic

	k          *poolmanager.Keeper
	gammKeeper types.PoolModuleI
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), poolmanager.NewMsgServerImpl(am.k))
	queryproto.RegisterQueryServer(cfg.QueryServer(), grpc.Querier{Q: pmclient.NewQuerier(am.k)})
	queryprotov2.RegisterQueryServer(cfg.QueryServer(), grpcv2.Querier{Q: pmclient.NewV2Querier(*am.k)})
}

func NewAppModule(poolmanagerKeeper poolmanager.Keeper, gammKeeper types.PoolModuleI) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		k:              &poolmanagerKeeper,
		gammKeeper:     gammKeeper,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// IsOnePerModuleType is a marker function just indicates that this is a one-per-module type.
func (am AppModule) IsOnePerModuleType() {}

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
}

// QuerierRoute returns the gamm module's querier route name.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// InitGenesis performs genesis initialization for the poolmanager module.
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genesisState types.GenesisState

	cdc.MustUnmarshalJSON(gs, &genesisState)

	am.k.InitGenesis(ctx, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the poolmanager.
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.k.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// BeginBlock performs cache initialization for the poolmanager module.
func (am AppModule) BeginBlock(context context.Context) error {
	ctx := sdk.UnwrapSDKContext(context)
	am.k.BeginBlock(ctx)
	return nil
}

// EndBlock performs alloy pool state updates for the poolmanager module.
func (am AppModule) EndBlock(context context.Context) error {
	ctx := sdk.UnwrapSDKContext(context)
	am.k.EndBlock(ctx)
	return nil
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// **** simulation implementation ****
// GenerateGenesisState creates a randomized GenState of the poolmanager module.
func (am AppModule) SimulatorGenesisState(simState *module.SimulationState, s *simtypes.SimCtx) {
	poolmanagerGen := types.DefaultGenesis()
	// change the pool creation fee denom from uosmo to stake
	poolmanagerGen.Params.PoolCreationFee = sdk.NewCoins(gammsimulation.PoolCreationFee)
	DefaultGenJson := simState.Cdc.MustMarshalJSON(poolmanagerGen)
	simState.GenState[types.ModuleName] = DefaultGenJson
}

func (am AppModule) Actions() []simtypes.Action {
	return nil
}
