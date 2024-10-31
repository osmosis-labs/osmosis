/*
Package gamm contains a variety of generalized automated market maker
functionality which provides the logic to create and interact with
liquidity pools on the Osmosis DEX.
 - Has pool creation, join pool, and exit pool logic
 - Token swap logic
 - GAMM pool queries
*/

package gamm

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
	"github.com/osmosis-labs/osmosis/v27/x/gamm/client/cli"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/stableswap"
	simulation "github.com/osmosis-labs/osmosis/v27/x/gamm/simulation"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/v2types"
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

// Name returns the gamm module's name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the gamm module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
	balancer.RegisterLegacyAminoCodec(cdc)
	stableswap.RegisterLegacyAminoCodec(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the gamm
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the gamm module.
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
	types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))     //nolint:errcheck
	v2types.RegisterQueryHandlerClient(context.Background(), mux, v2types.NewQueryClient(clientCtx)) //nolint:errcheck
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
	balancer.RegisterInterfaces(registry)
	stableswap.RegisterInterfaces(registry)
}

type AppModule struct {
	AppModuleBasic

	ak     types.AccountKeeper
	bk     types.BankKeeper
	keeper keeper.Keeper
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(&am.keeper))
	balancer.RegisterMsgServer(cfg.MsgServer(), keeper.NewBalancerMsgServerImpl(&am.keeper))
	stableswap.RegisterMsgServer(cfg.MsgServer(), keeper.NewStableswapMsgServerImpl(&am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerier(am.keeper))
	v2types.RegisterQueryServer(cfg.QueryServer(), keeper.NewV2Querier(am.keeper))
}

func NewAppModule(cdc codec.Codec, keeper keeper.Keeper,
	accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		ak:             accountKeeper,
		bk:             bankKeeper,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// IsOnePerModuleType is a marker function just indicates that this is a one-per-module type.
func (am AppModule) IsOnePerModuleType() {}

// RegisterInvariants registers the gamm module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper, am.bk)
}

// QuerierRoute returns the gamm module's querier route name.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// InitGenesis performs genesis initialization for the gamm module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)
	am.keeper.InitGenesis(ctx, genState, am.cdc)
}

// ExportGenesis returns the exported genesis state as raw bytes for the gamm
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// **** simulation implementation ****
// GenerateGenesisState creates a randomized GenState of the gamm module.
func (am AppModule) SimulatorGenesisState(simState *module.SimulationState, s *simtypes.SimCtx) {
	DefaultGen := types.DefaultGenesis()
	DefaultGenJson := simState.Cdc.MustMarshalJSON(DefaultGen)
	simState.GenState[types.ModuleName] = DefaultGenJson
}

func (am AppModule) Actions() []simtypes.Action {
	return simulation.DefaultActions(am.keeper)
}
