package v33

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v32/app/keepers"
	"github.com/osmosis-labs/osmosis/v32/app/upgrades"
	v3 "github.com/osmosis-labs/osmosis/v32/x/cosmwasmpool/cosmwasm/msg/v3"
	incentiveskeeper "github.com/osmosis-labs/osmosis/v32/x/incentives/keeper"
)

// CreateUpgradeHandler returns the v33 upgrade handler. After the module migrations it recovers the
// HUAHUA held by gauges 1954-1959 (see HuahuaStuckGauges) and, on mainnet, migrates native WBTC
// holders to allBTC (see MigrateWBTCToAllBTC).
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)

		// The recovery is not critical to the chain, so a failure is logged instead of halting the upgrade.
		// It runs in a cache context so that it is applied either entirely or not at all.
		cacheCtx, write := sdkCtx.CacheContext()
		recovered, err := recoverHuahuaFromStuckGauges(cacheCtx, keepers.IncentivesKeeper)
		if err != nil {
			sdkCtx.Logger().Error("v33 upgrade: HUAHUA recovery skipped", "error", err)
		} else {
			write()
			sdkCtx.Logger().Info("v33 upgrade: HUAHUA recovered", "amount", recovered.String(), "recipient", HuahuaRecoveryAddress)
		}

		if sdkCtx.ChainID() == MainnetChainID {
			migrateWBTCAndLog(sdkCtx, keepers, AllBTCTransmuterAddress, WBTCDenom, AllBTCDenom)
		} else {
			sdkCtx.Logger().Info("v33 upgrade: skipping WBTC to allBTC migration on non-mainnet chain", "chain_id", sdkCtx.ChainID())
		}

		return migrations, nil
	}
}

// recoverHuahuaFromStuckGauges finishes the stuck HUAHUA gauges and sends their undistributed coins
// to the recovery address. See HuahuaStuckGauges.
//
// All undistributed coins of each gauge are sent, not only HUAHUA. On mainnet these gauges hold only
// HUAHUA, but MsgAddToGauge can add other denoms before the upgrade; once the gauge is finished those
// coins could never be distributed or withdrawn, so they are sent along with the HUAHUA.
func recoverHuahuaFromStuckGauges(ctx sdk.Context, incentivesKeeper *incentiveskeeper.Keeper) (sdk.Coins, error) {
	recipient, err := sdk.AccAddressFromBech32(HuahuaRecoveryAddress)
	if err != nil {
		return nil, err
	}

	recovered := sdk.NewCoins()
	for _, stuck := range HuahuaStuckGauges {
		gauge, err := incentivesKeeper.GetGaugeByID(ctx, stuck.GaugeID)
		if err != nil {
			return nil, err
		}
		// Guard against the IDs pointing to anything other than the expected gauges.
		if gauge.DistributeTo.Denom != stuck.Denom {
			return nil, fmt.Errorf("gauge %d distributes to %q, expected %q", stuck.GaugeID, gauge.DistributeTo.Denom, stuck.Denom)
		}
		if gauge.Coins.AmountOf(HuahuaDenom).IsZero() {
			return nil, fmt.Errorf("gauge %d holds no HUAHUA", stuck.GaugeID)
		}

		coins, err := incentivesKeeper.ForceFinishGaugeAndSendUndistributed(ctx, stuck.GaugeID, recipient)
		if err != nil {
			return nil, err
		}
		recovered = recovered.Add(coins...)
	}
	return recovered, nil
}

// WBTCMigrationResult summarizes a WBTC to allBTC migration.
type WBTCMigrationResult struct {
	// Migrated is the number of accounts whose WBTC was deposited into the transmuter.
	Migrated int
	// Skipped is the number of holders that are not user accounts or have no spendable WBTC.
	Skipped int
	// Failed is the number of user accounts whose deposit failed and was rolled back.
	Failed int
	// Total is the amount of WBTC deposited into the transmuter.
	Total osmomath.Int
}

// migrateWBTCAndLog runs MigrateWBTCToAllBTC and logs the outcome. The migration is not critical to
// the chain, so a failure is logged instead of halting the upgrade.
func migrateWBTCAndLog(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter, wbtcDenom, alloyedDenom string) {
	result, err := MigrateWBTCToAllBTC(ctx, keepers, transmuter, wbtcDenom, alloyedDenom)
	if err != nil {
		ctx.Logger().Error("v33 upgrade: WBTC to allBTC migration skipped", "error", err)
		return
	}
	ctx.Logger().Info("v33 upgrade: WBTC migrated to allBTC",
		"migrated", result.Migrated, "skipped", result.Skipped, "failed", result.Failed,
		"total", sdk.NewCoin(wbtcDenom, result.Total).String())
}

// MigrateWBTCToAllBTC deposits the spendable wbtcDenom balance of every user account into the alloy
// transmuter with join_pool, executed as the account, so the account receives alloyedDenom in return.
// Module accounts and wasm contracts are skipped. Each deposit runs in its own cache context: a failed
// deposit is logged and rolled back without affecting the others.
//
// An error is returned, with no state changed, if the transmuter does not accept wbtcDenom or there
// are more than MaxWBTCHoldersToMigrate holders.
func MigrateWBTCToAllBTC(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter, wbtcDenom, alloyedDenom string) (WBTCMigrationResult, error) {
	result := WBTCMigrationResult{Total: osmomath.ZeroInt()}
	transmuterAddr, err := sdk.AccAddressFromBech32(transmuter)
	if err != nil {
		return result, fmt.Errorf("transmuter address: %w", err)
	}
	if err := checkTransmuterAcceptsDenom(ctx, keepers, transmuterAddr, wbtcDenom); err != nil {
		return result, err
	}
	// Collect the holders before changing any balance, since every deposit updates the denom index.
	holders, err := collectHolders(ctx, keepers, wbtcDenom, MaxWBTCHoldersToMigrate)
	if err != nil {
		return result, err
	}

	for _, holder := range holders {
		if !isUserAccount(ctx, keepers, holder) {
			result.Skipped++
			continue
		}
		coin := keepers.BankKeeper.SpendableCoin(ctx, holder, wbtcDenom)
		if !coin.IsPositive() {
			result.Skipped++
			continue
		}
		err := osmoutils.ApplyFuncIfNoErrorLogToDebug(ctx, func(cacheCtx sdk.Context) error {
			return joinAsHolder(cacheCtx, keepers, transmuterAddr, holder, coin, alloyedDenom)
		})
		if err != nil {
			ctx.Logger().Error("v33 upgrade: WBTC migration failed for account", "address", holder.String(), "amount", coin.String(), "error", err)
			result.Failed++
			continue
		}
		result.Migrated++
		result.Total = result.Total.Add(coin.Amount)
	}
	return result, nil
}

// checkTransmuterAcceptsDenom returns an error unless transmuter is a contract that lists denom in
// its asset configs.
func checkTransmuterAcceptsDenom(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter sdk.AccAddress, denom string) error {
	if !keepers.WasmKeeper.HasContractInfo(ctx, transmuter) {
		return fmt.Errorf("transmuter %s is not a contract", transmuter)
	}
	respBz, err := keepers.WasmKeeper.QuerySmart(ctx, transmuter, []byte(`{"list_asset_configs": {}}`))
	if err != nil {
		return fmt.Errorf("querying asset configs of transmuter %s: %w", transmuter, err)
	}
	var resp v3.ListAssetConfigsResponse
	if err := json.Unmarshal(respBz, &resp); err != nil {
		return fmt.Errorf("decoding asset configs of transmuter %s: %w", transmuter, err)
	}
	for _, config := range resp.AssetConfigs {
		if config.Denom == denom {
			return nil
		}
	}
	return fmt.Errorf("transmuter %s does not accept %s", transmuter, denom)
}

// collectHolders returns the addresses with a balance of denom, in address order. It returns an error
// if there are more than maxHolders.
func collectHolders(ctx sdk.Context, keepers *keepers.AppKeepers, denom string, maxHolders int) ([]sdk.AccAddress, error) {
	iter, err := keepers.BankKeeper.Balances.Indexes.Denom.MatchExact(ctx, denom)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	holders := make([]sdk.AccAddress, 0, 1024)
	for ; iter.Valid(); iter.Next() {
		if len(holders) == maxHolders {
			return nil, fmt.Errorf("more than %d holders of %s", maxHolders, denom)
		}
		pk, err := iter.PrimaryKey()
		if err != nil {
			return nil, err
		}
		holders = append(holders, pk.K1())
	}
	return holders, nil
}

// isUserAccount reports whether addr is an existing account that is neither a module account
// (which includes GAMM and CL pools) nor a wasm contract (which includes cosmwasm pools).
func isUserAccount(ctx sdk.Context, keepers *keepers.AppKeepers, addr sdk.AccAddress) bool {
	acc := keepers.AccountKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return false
	}
	if _, isModule := acc.(sdk.ModuleAccountI); isModule {
		return false
	}
	return !keepers.WasmKeeper.HasContractInfo(ctx, addr)
}

// joinAsHolder deposits coin into the transmuter on behalf of holder and checks that holder spent
// exactly coin and received some alloyedDenom in return.
func joinAsHolder(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter, holder sdk.AccAddress, coin sdk.Coin, alloyedDenom string) error {
	wbtcBefore := keepers.BankKeeper.GetBalance(ctx, holder, coin.Denom).Amount
	alloyedBefore := keepers.BankKeeper.GetBalance(ctx, holder, alloyedDenom).Amount

	if _, err := keepers.ContractKeeper.Execute(ctx, transmuter, holder, []byte(`{"join_pool": {}}`), sdk.NewCoins(coin)); err != nil {
		return err
	}

	if spent := wbtcBefore.Sub(keepers.BankKeeper.GetBalance(ctx, holder, coin.Denom).Amount); !spent.Equal(coin.Amount) {
		return fmt.Errorf("spent %s %s, expected %s", spent, coin.Denom, coin.Amount)
	}
	if !keepers.BankKeeper.GetBalance(ctx, holder, alloyedDenom).Amount.GT(alloyedBefore) {
		return errors.New("no alloyed asset received")
	}
	return nil
}
