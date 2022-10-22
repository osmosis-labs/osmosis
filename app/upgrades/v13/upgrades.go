package v13

import (
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	"github.com/osmosis-labs/osmosis/v12/app/keepers"
	"github.com/osmosis-labs/osmosis/v12/app/upgrades"
	ratelimittypes "github.com/osmosis-labs/osmosis/v12/x/ibc-rate-limit/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v12/x/lockup/types"
)

func setupRateLimiting(ctx sdk.Context, keepers *keepers.AppKeepers) error {

	govModule := keepers.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
	code := []byte{} // ToDo: read the code
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(keepers.WasmKeeper)
	instantiateConfig := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeOnlyAddress, Address: govModule.String()}
	codeID, err := contractKeeper.Create(ctx, govModule, code, &instantiateConfig)
	if err != nil {
		return err
	}

	transferModule := keepers.AccountKeeper.GetModuleAddress(transfertypes.ModuleName)

	// ToDo: Configure quotas

	//      quotas := fmt.Sprintf(`
	//      {"channel_id": "channel-0", "denom": "%s", "quotas": [{"name":"%s", "duration": %d, "send_recv":[%d, %d]}] }
	//`, sdk.DefaultBondDenom, name, duration, send_precentage, recv_percentage)
	quotas := "{}"

	initMsgBz := []byte(fmt.Sprintf(`{
           "gov_module":  "%s",
           "ibc_module":"%s",
           "paths": [%s]
        }`,
		govModule, transferModule, quotas))

	addr, _, err := contractKeeper.Instantiate(ctx, codeID, govModule, govModule, initMsgBz, "rate limiting contract", nil)
	if err != nil {
		return err
	}
	addrStr, err := sdk.Bech32ifyAddressBytes("osmo", addr)
	params, err := ratelimittypes.NewParams(addrStr)
	paramSpace, ok := keepers.ParamsKeeper.GetSubspace(ratelimittypes.ModuleName)
	if !ok {
		return sdkerrors.New("rate-limiting-upgrades", 2, "can't create paramspace")
	}
	paramSpace.SetParamSet(ctx, &params)
	return nil
}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		keepers.LockupKeeper.SetParams(ctx, lockuptypes.DefaultParams())

		setupRateLimiting(ctx, keepers)

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
