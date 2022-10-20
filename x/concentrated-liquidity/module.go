package concentrated_liquidity

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/v12/x/twap/client/queryproto"
	"github.com/osmosis-labs/osmosis/v12/x/twap/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
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
	return nil
}

// ---------------------------------------
// Interfaces.
func (b AppModuleBasic) RegisterRESTRoutes(ctx client.Context, r *mux.Router) {
}

func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	queryproto.RegisterQueryHandlerClient(context.Background(), mux, queryproto.NewQueryClient(clientCtx)) //nolint:errcheck
}

func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// RegisterInterfaces registers interfaces and implementations of the gamm module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
}

type AppModule struct {
	AppModuleBasic

	k Keeper
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
}

func NewAppModule(concentratedLiquidityKeeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		k:              Keeper{},
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

// InitGenesis performs genesis initialization for the twap module.
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the twap.
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return []byte{}
}

// BeginBlock performs a no-op.
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock performs a no-op.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }
