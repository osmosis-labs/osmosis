package ibcratelimitmodule

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	"cosmossdk.io/core/appmodule"

	ibcratelimit "github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit"
	ibcratelimitclient "github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/client"
	ibcratelimitcli "github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/client/cli"
	"github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/client/grpc"
	"github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/types"
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

// ValidateGenesis performs genesis state validation for the ibc-rate-limit module.
func (b AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
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
}

func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	return ibcratelimitcli.GetQueryCmd()
}

// RegisterInterfaces registers interfaces and implementations of the ibc-rate-limit module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface for the capability module.
type AppModule struct {
	AppModuleBasic

	ics4wrapper ibcratelimit.ICS4Wrapper
}

func NewAppModule(ics4wrapper ibcratelimit.ICS4Wrapper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		ics4wrapper:    ics4wrapper,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// IsOnePerModuleType is a marker function just indicates that this is a one-per-module type.
func (am AppModule) IsOnePerModuleType() {}

// Name returns the txfees module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// QuerierRoute returns the ibc-rate-limit module's query routing key.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	queryproto.RegisterQueryServer(cfg.QueryServer(), grpc.Querier{Q: ibcratelimitclient.Querier{K: am.ics4wrapper}})
}

// RegisterInvariants registers the txfees module's invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs the txfees module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)
	am.ics4wrapper.InitGenesis(ctx, genState)
}

// ExportGenesis returns the txfees module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.ics4wrapper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }
