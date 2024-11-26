package downtimemodule

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	downtimedetector "github.com/osmosis-labs/osmosis/v27/x/downtime-detector"
	downtimeclient "github.com/osmosis-labs/osmosis/v27/x/downtime-detector/client"
	downtimecli "github.com/osmosis-labs/osmosis/v27/x/downtime-detector/client/cli"
	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/client/grpc"
	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/types"
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
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	queryproto.RegisterQueryHandlerClient(context.Background(), mux, queryproto.NewQueryClient(clientCtx)) //nolint:errcheck
}

func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	return downtimecli.GetQueryCmd()
}

func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
}

type AppModule struct {
	AppModuleBasic

	k downtimedetector.Keeper
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	queryproto.RegisterQueryServer(cfg.QueryServer(), grpc.Querier{Q: downtimeclient.Querier{K: am.k}})
}

func NewAppModule(k downtimedetector.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		k:              k,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// IsOnePerModuleType is a marker function just indicates that this is a one-per-module type.
func (am AppModule) IsOnePerModuleType() {}

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

func (AppModule) QuerierRoute() string { return types.RouterKey }

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genesisState types.GenesisState

	cdc.MustUnmarshalJSON(gs, &genesisState)

	am.k.InitGenesis(ctx, &genesisState)
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.k.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

func (am AppModule) BeginBlock(context context.Context) error {
	ctx := sdk.UnwrapSDKContext(context)
	am.k.BeginBlock(ctx)
	return nil
}

func (AppModule) ConsensusVersion() uint64 { return 1 }
