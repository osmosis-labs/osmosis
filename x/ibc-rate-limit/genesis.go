package ibc_rate_limit

import (
	"embed"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/osmosis-labs/osmosis/v15/x/ibc-rate-limit/types"
)

// InitGenesis initializes the x/ibc-rate-limit module's state from a provided genesis
// state, which includes the parameter for the contract address.
func (i *ICS4Wrapper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	err := SetupRateLimiting(ctx, i.accountKeeper, i.ContractKeeper, i.paramSpace)
	if err != nil {
		panic(err)
	}
	i.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the x/ibc-rate-limit module's exported genesis.
func (i *ICS4Wrapper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: i.GetParams(ctx),
	}
}

//go:embed bytecode/rate_limiter.wasm
var embedFs embed.FS

func SetupRateLimiting(ctx sdk.Context, accountKeeper *authkeeper.AccountKeeper, contractKeeper *wasmkeeper.PermissionedKeeper, paramSpace paramtypes.Subspace) error {
	govModule := accountKeeper.GetModuleAddress(govtypes.ModuleName)
	code, err := embedFs.ReadFile("bytecode/rate_limiter.wasm")
	if err != nil {
		return err
	}
	instantiateConfig := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeOnlyAddress, Address: govModule.String()}
	codeID, _, err := contractKeeper.Create(ctx, govModule, code, &instantiateConfig)
	if err != nil {
		return err
	}

	transferModule := accountKeeper.GetModuleAddress(transfertypes.ModuleName)

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
	paramSpace.SetParamSet(ctx, &params)
	return nil
}
