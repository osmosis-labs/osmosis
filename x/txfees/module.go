/*
The txfees modules allows nodes to easily support many
tokens for usage as txfees, while letting node operators
only specify their tx fee parameters for a single "base" asset.

- Adds a whitelist of tokens that can be used as fees on the chain.
- Any token not on this list cannot be provided as a tx fee.
- Adds a new SDK message for creating governance proposals for adding new TxFee denoms.
*/
package txfees

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/txfees/client/cli"
	"github.com/osmosis-labs/osmosis/v27/x/txfees/keeper"
	mempool1559 "github.com/osmosis-labs/osmosis/v27/x/txfees/keeper/mempool-1559"
	"github.com/osmosis-labs/osmosis/v27/x/txfees/types"
)

var (
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ module.HasGenesisBasics = AppModuleBasic{}

	_ appmodule.AppModule        = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}

	cachedConsParams cmtproto.ConsensusParams
)

const ModuleName = types.ModuleName

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic implements the AppModuleBasic interface for the txfees module.
type AppModuleBasic struct{}

func NewAppModuleBasic() AppModuleBasic {
	return AppModuleBasic{}
}

// Name returns the txfees module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types.
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis returns the txfees module's default genesis state.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the txfee module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	//nolint:errcheck
	types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

// GetQueryCmd returns the txfees module's root query command.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface for the txfees module.
type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

func NewAppModule(keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(),
		keeper:         keeper,
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

// QuerierRoute returns the txfees module's query routing key.
func (AppModule) QuerierRoute() string { return "" }

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerier(am.keeper))
}

// RegisterInvariants registers the txfees module's invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs the txfees module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)

	am.keeper.InitGenesis(ctx, genState)
}

// ExportGenesis returns the txfees module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// BeginBlock executes all ABCI BeginBlock logic respective to the txfees module.
func (am AppModule) BeginBlock(context context.Context) error {
	ctx := sdk.UnwrapSDKContext(context)
	mempool1559.BeginBlockCode(ctx)

	// Check if the block gas limit has changed.
	// If it has, update the target gas for eip1559.
	am.CheckAndSetTargetGas(ctx)
	return nil
}

// EndBlock executes all ABCI EndBlock logic respective to the txfees module. It
// returns no validator updates.
func (am AppModule) EndBlock(context context.Context) error {
	ctx := sdk.UnwrapSDKContext(context)
	mempool1559.EndBlockCode(ctx)
	return nil
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// On start, we unmarshal the consensus params once and cache them.
// Then, on every block, we check if the current consensus param bytes have changed in comparison to the cached value.
// If they have, we unmarshal the current consensus params, update the target gas, and cache the value.
// This is done to improve performance by not having to fetch and unmarshal the consensus params on every block.
// TODO: Move this to EIP-1559 code
func (am AppModule) CheckAndSetTargetGas(ctx sdk.Context) {
	// Check if the block gas limit has changed.
	// If it has, update the target gas for eip1559.
	consParams, err := am.keeper.GetConsParams(ctx)
	if err != nil {
		panic(err)
	}

	// If cachedConsParams is empty, set equal to consParams and set the target gas.
	if cachedConsParams.Equal(cmtproto.ConsensusParams{}) {
		cachedConsParams = *consParams.Params

		// Check if cachedConsParams.Block is nil to prevent panic
		if cachedConsParams.Block == nil || cachedConsParams.Block.MaxGas == 0 {
			return
		}

		if cachedConsParams.Block.MaxGas == -1 {
			return
		}

		newBlockMaxGas := mempool1559.TargetBlockSpacePercent.Mul(osmomath.NewDec(cachedConsParams.Block.MaxGas)).TruncateInt().Int64()
		mempool1559.TargetGas = newBlockMaxGas
		return
	}

	// If the consensus params have changed, check if it was maxGas that changed. If so, update the target gas.
	if consParams.Params.Block.MaxGas != cachedConsParams.Block.MaxGas {
		if consParams.Params.Block.MaxGas == -1 {
			return
		}

		newBlockMaxGas := mempool1559.TargetBlockSpacePercent.Mul(osmomath.NewDec(consParams.Params.Block.MaxGas)).TruncateInt().Int64()
		mempool1559.TargetGas = newBlockMaxGas
		cachedConsParams = *consParams.Params
	}
}
