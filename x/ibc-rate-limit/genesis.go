package ibc_rate_limit

import (
	"embed"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"

	"github.com/osmosis-labs/osmosis/v14/x/ibc-rate-limit/types"
)

// InitGenesis initializes the x/ibc-rate-limit module's state from a provided genesis
// state, which includes the current live pools, global pool parameters (e.g. pool creation fee), next pool id etc.
// TODO: test
func (i *ICS4Wrapper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	i.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the x/ibc-rate-limit module's exported genesis.
// TODO: test
func (i *ICS4Wrapper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params, _ := i.GetParams(ctx)
	return &types.GenesisState{
		Params: params,
	}
}

//go:embed bytecode/rate_limiter.wasm
var EmbedFs embed.FS

func (i *ICS4Wrapper) InitContract(ctx sdk.Context, wasmKeeper *wasmkeeper.Keeper) error {
	govModule := i.accountKeeper.GetModuleAddress(govtypes.ModuleName)
	code, err := EmbedFs.ReadFile("bytecode/rate_limiter.wasm")
	if err != nil {
		return err
	}
	contractKeeper := wasmkeeper.NewGovPermissionKeeper(wasmKeeper)
	instantiateConfig := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeOnlyAddress, Address: govModule.String()}
	codeID, _, err := contractKeeper.Create(ctx, govModule, code, &instantiateConfig)
	if err != nil {
		return err
	}
	transferModule := i.accountKeeper.GetModuleAddress(transfertypes.ModuleName)

	initMsgBz := []byte(fmt.Sprintf(`{
           "gov_module":  "%s",
           "ibc_module":"%s",
           "paths": []
        }`,
		govModule, transferModule))

	addr, _, err := contractKeeper.Instantiate(ctx, codeID, govModule, govModule, initMsgBz, "rate limiting contract", nil)
	if err != nil {
		return err
	}
	addrStr, err := sdk.Bech32ifyAddressBytes("osmo", addr)
	if err != nil {
		return err
	}
	params, err := types.NewParams(addrStr)
	if err != nil {
		return err
	}
	i.SetParams(ctx, params)
	return nil

}
