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
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v20/simulation/simtypes"
	gammsimulation "github.com/osmosis-labs/osmosis/v20/x/gamm/simulation"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager"
	pmclient "github.com/osmosis-labs/osmosis/v20/x/poolmanager/client"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/client/cli"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/client/grpc"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/client/grpcv2"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/client/queryproto"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/client/queryprotov2"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
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
func (b AppModuleBasic) RegisterRESTRoutes(ctx client.Context, r *mux.Router) {
}

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

	k          poolmanager.Keeper
	gammKeeper types.PoolModuleI
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), poolmanager.NewMsgServerImpl(&am.k))
	queryproto.RegisterQueryServer(cfg.QueryServer(), grpc.Querier{Q: pmclient.NewQuerier(am.k)})
	queryprotov2.RegisterQueryServer(cfg.QueryServer(), grpcv2.Querier{Q: pmclient.NewV2Querier(am.k)})
}

func NewAppModule(poolmanagerKeeper poolmanager.Keeper, gammKeeper types.PoolModuleI) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		k:              poolmanagerKeeper,
		gammKeeper:     gammKeeper,
	}
}

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
}

func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// QuerierRoute returns the gamm module's querier route name.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// LegacyQuerierHandler returns the x/gamm module's sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(sdk.Context, []string, abci.RequestQuery) ([]byte, error) {
		return nil, fmt.Errorf("legacy querier not supported for the x/%s module", types.ModuleName)
	}
}

// InitGenesis performs genesis initialization for the poolmanager module.
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState

	cdc.MustUnmarshalJSON(gs, &genesisState)

	am.k.InitGenesis(ctx, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the poolmanager.
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.k.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// BeginBlock performs a no-op.
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock performs a no-op.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
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
