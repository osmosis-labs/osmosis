package module

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v27/simulation/simtypes"
	cosmwasmpool "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool"
	moduleclient "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/client"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/client/cli"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/client/grpc"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
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
	model.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the cosmwasmpool module.
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
}

func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.NewQueryCmd()
}

// RegisterInterfaces registers interfaces and implementations of the gamm module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
	model.RegisterInterfaces(registry)
}

type AppModule struct {
	AppModuleBasic

	k cosmwasmpool.Keeper
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), cosmwasmpool.NewMsgServerImpl(&am.k))
	model.RegisterMsgCreatorServer(cfg.MsgServer(), cosmwasmpool.NewMsgCreatorServerImpl(&am.k))
	queryproto.RegisterQueryServer(cfg.QueryServer(), grpc.Querier{Q: moduleclient.NewQuerier(am.k)})
}

func NewAppModule(cdc codec.Codec, cosmwasmpoolKeeper cosmwasmpool.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		k:              cosmwasmpoolKeeper,
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

// InitGenesis performs genesis initialization for the cosmwasmpool module.
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genState)
	am.k.InitGenesis(ctx, &genState, am.cdc)
}

// ExportGenesis returns the exported genesis state as raw bytes for the cosmwasmpool.
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.k.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

func (am AppModule) Actions() []simtypes.Action {
	return nil
}
