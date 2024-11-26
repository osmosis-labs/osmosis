package concentrated_liquidity_module

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/v27/simulation/simtypes"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client/cli"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client/queryproto"
	clmodel "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/simulation"

	clkeeper "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	clclient "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client/grpc"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types/genesis"
)

var (
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ module.HasGenesisBasics = AppModuleBasic{}

	_ appmodule.AppModule        = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}
)

type AppModuleBasic struct {
	cdc codec.Codec
}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterCodec(cdc)
	clmodel.RegisterCodec(cdc)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(genesis.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the cl module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState genesis.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// ---------------------------------------
// Interfaces.
func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	queryproto.RegisterQueryHandlerClient(context.Background(), mux, queryproto.NewQueryClient(clientCtx)) //nolint:errcheck
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
	clmodel.RegisterInterfaces(registry)
}

type AppModule struct {
	AppModuleBasic

	keeper clkeeper.Keeper
}

func NewAppModule(cdc codec.Codec, keeper clkeeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// IsOnePerModuleType is a marker function just indicates that this is a one-per-module type.
func (am AppModule) IsOnePerModuleType() {}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), clkeeper.NewMsgServerImpl(&am.keeper))
	clmodel.RegisterMsgServer(cfg.MsgServer(), clkeeper.NewMsgCreatorServerImpl(&am.keeper))
	queryproto.RegisterQueryServer(cfg.QueryServer(), grpc.Querier{Q: clclient.Querier{Keeper: am.keeper}})
}

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
}

// QuerierRoute returns the gamm module's querier route name.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// InitGenesis performs genesis initialization for the cl module.
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genState genesis.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)

	am.keeper.InitGenesis(ctx, genState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the twap.
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// ___________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the valset module.
func (am AppModule) GenerateGenesisState(simState *module.SimulationState, s *simtypes.SimCtx) {
}

func (am AppModule) Actions() []simtypes.Action {
	return []simtypes.Action{
		simtypes.NewMsgBasedAction("CreateConcentratedPool", am.keeper, simulation.RandomMsgCreateConcentratedPool),
		simtypes.NewMsgBasedAction("CreatePosition", am.keeper, simulation.RandMsgCreatePosition),
		simtypes.NewMsgBasedAction("WithdrawPosition", am.keeper, simulation.RandMsgWithdrawPosition),
		simtypes.NewMsgBasedAction("CollectSpreadRewards", am.keeper, simulation.RandMsgCollectSpreadRewards),
		simtypes.NewMsgBasedAction("CollectIncentives", am.keeper, simulation.RandMsgCollectIncentives),
	}
}
