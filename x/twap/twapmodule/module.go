package twapmodule

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

	"github.com/osmosis-labs/osmosis/v27/x/twap"
	twapclient "github.com/osmosis-labs/osmosis/v27/x/twap/client"
	twapcli "github.com/osmosis-labs/osmosis/v27/x/twap/client/cli"
	"github.com/osmosis-labs/osmosis/v27/x/twap/client/grpc"
	"github.com/osmosis-labs/osmosis/v27/x/twap/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types"
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
	queryproto.RegisterQueryHandlerClient(context.Background(), mux, queryproto.NewQueryClient(clientCtx)) //nolint:errcheck
}

func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
	// return cli.NewTxCmd()
}

func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	return twapcli.GetQueryCmd()
}

// RegisterInterfaces registers interfaces and implementations of the gamm module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
}

type AppModule struct {
	AppModuleBasic

	k twap.Keeper
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	queryproto.RegisterQueryServer(cfg.QueryServer(), grpc.Querier{Q: twapclient.Querier{K: am.k}})
}

func NewAppModule(twapKeeper twap.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		k:              twapKeeper,
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

// InitGenesis performs genesis initialization for the twap module.
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genesisState types.GenesisState

	cdc.MustUnmarshalJSON(gs, &genesisState)

	am.k.InitGenesis(ctx, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the twap.
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.k.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// EndBlock executes all ABCI EndBlock logic respective to the TWAP module. It
// returns no validator updates.
func (am AppModule) EndBlock(context context.Context) error {
	ctx := sdk.UnwrapSDKContext(context)
	am.k.EndBlock(ctx)
	return nil
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }
