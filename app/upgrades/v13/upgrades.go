package v13

import (
	"fmt"
	"os"
	"strings"

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
	ibcratelimittypes "github.com/osmosis-labs/osmosis/v12/x/ibc-rate-limit/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v12/x/lockup/types"
)

func setupRateLimiting(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	govModule := keepers.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
	code, err := os.ReadFile("rate_limiter.wasm")
	if err != nil {
		return err
	}
	contractKeeper := wasmkeeper.NewGovPermissionKeeper(keepers.WasmKeeper)
	instantiateConfig := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeOnlyAddress, Address: govModule.String()}
	codeID, err := contractKeeper.Create(ctx, govModule, code, &instantiateConfig)
	if err != nil {
		return err
	}

	transferModule := keepers.AccountKeeper.GetModuleAddress(transfertypes.ModuleName)

	denoms := []string{
		"uosmo",
		"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", // atom
		"ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", // usdc
		"ibc/EA1D43981D5C9A1C4AAEA9C23BB1D4FA126BA9BC7020A25E0AE4AA841EA25DC5", // weth
		"ibc/EA1D43981D5C9A1C4AAEA9C23BB1D4FA126BA9BC7020A25E0AE4AA841EA25DC5", // wbtc
	}

	quotas := ""
	for _, denom := range denoms {
		quotas += fmt.Sprintf(`
{"channel_id": "any", "denom": "%s", "quotas": [{"name":"weekly", "duration": 604800, "send_recv":[50, 50]}] },`, denom)
	}
	if quotas != "" {
		// Remove the trailing comma because JSON can't handle it
		quotas = strings.TrimRight(quotas, ",")
	}

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
	if err != nil {
		return err
	}
	params, err := ibcratelimittypes.NewParams(addrStr)
	if err != nil {
		return err
	}
	paramSpace, ok := keepers.ParamsKeeper.GetSubspace(ibcratelimittypes.ModuleName)
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
		if err := setupRateLimiting(ctx, keepers); err != nil {
			return nil, err
		}
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
