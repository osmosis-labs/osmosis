package v33

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/osmosis-labs/osmosis/v32/app/keepers"
	appparams "github.com/osmosis-labs/osmosis/v32/app/params"
	"github.com/osmosis-labs/osmosis/v32/app/upgrades"
	incentiveskeeper "github.com/osmosis-labs/osmosis/v32/x/incentives/keeper"
)

// CreateUpgradeHandler returns the v33 upgrade handler. After the module migrations it recovers the
// HUAHUA held by gauges 1954-1959 (see HuahuaStuckGauges), raises every validator's commission to
// MinCommissionRate and, on mainnet, moves OSMO to the burn address and migrates native WBTC holders
// to allBTC (see MigrateWBTCToAllBTC).
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
		if sdkCtx.ChainID() == MainnetChainID {
			if err := moveOsmoToBurnAddress(sdkCtx, keepers); err != nil {
				return nil, err
			}
		} else {
			sdkCtx.Logger().Info("skipping v33 OSMO burn transfer on non-mainnet chain", "chain_id", sdkCtx.ChainID())
		}

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

		// Like the HUAHUA recovery, a failure is logged and the commission change is applied entirely or not at all.
		cacheCtx, write = sdkCtx.CacheContext()
		raised, err := raiseMinCommission(cacheCtx, keepers.StakingKeeper)
		if err != nil {
			sdkCtx.Logger().Error("v33 upgrade: minimum commission raise skipped", "error", err)
		} else {
			write()
			sdkCtx.Logger().Info("v33 upgrade: minimum commission raised", "min_commission_rate", MinCommissionRate.String(), "validators_modified", raised)
		}

		if sdkCtx.ChainID() == MainnetChainID {
			migrateWBTCAndLog(sdkCtx, keepers)
		} else {
			sdkCtx.Logger().Info("v33 upgrade: skipping WBTC to allBTC migration on non-mainnet chain", "chain_id", sdkCtx.ChainID())
		}

		return migrations, nil
	}
}

// moveOsmoToBurnAddress sends each source account's entire uosmo balance at upgrade
// time. Other denoms remain in the source accounts. Empty balances are skipped.
func moveOsmoToBurnAddress(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	to, err := sdk.AccAddressFromBech32(OsmoBurnAddress)
	if err != nil {
		return fmt.Errorf("v33 OSMO burn recipient: %w", err)
	}
	for _, source := range OsmoBurnSourceAddresses {
		from, err := sdk.AccAddressFromBech32(source)
		if err != nil {
			return fmt.Errorf("v33 OSMO burn source %s: %w", source, err)
		}
		balance := keepers.BankKeeper.GetBalance(ctx, from, appparams.BaseCoinUnit)
		if balance.IsZero() {
			continue
		}
		if err := keepers.BankKeeper.SendCoins(ctx, from, to, sdk.NewCoins(balance)); err != nil {
			return fmt.Errorf("v33 moving OSMO from %s to burn address: %w", source, err)
		}
		ctx.Logger().Info("v33 upgrade: OSMO moved to burn address", "from", source, "to", OsmoBurnAddress, "amount", balance.String())
	}
	return nil
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

// raiseMinCommission sets the staking MinCommissionRate to the v33 minimum and raises every validator
// (of any status) below it, as approved by governance in proposal 1066. It returns the number of
// validators modified.
//
// MinCommissionRate is only enforced when a validator is created or edits its commission, so the
// parameter change alone would leave existing rates unchanged. The rates are therefore rewritten
// directly, skipping Commission.ValidateNewRate: its 24h and max_change_rate limits protect delegators
// from sudden increases by an operator, whereas this increase is mandated by governance.
// A validator's MaxRate below the minimum is raised to it as well. MaxChangeRate needs no change:
// it is at most the old MaxRate, so it stays within the new one.
func raiseMinCommission(ctx sdk.Context, stakingKeeper *stakingkeeper.Keeper) (int, error) {
	params, err := stakingKeeper.GetParams(ctx)
	if err != nil {
		return 0, err
	}
	params.MinCommissionRate = MinCommissionRate
	if err := stakingKeeper.SetParams(ctx, params); err != nil {
		return 0, err
	}

	validators, err := stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return 0, err
	}
	modified := 0
	for _, validator := range validators {
		oldCommission := validator.Commission
		if oldCommission.Rate.GTE(MinCommissionRate) && oldCommission.MaxRate.GTE(MinCommissionRate) {
			continue
		}

		newCommission := oldCommission
		if newCommission.MaxRate.LT(MinCommissionRate) {
			newCommission.MaxRate = MinCommissionRate
		}
		if newCommission.Rate.LT(MinCommissionRate) {
			newCommission.Rate = MinCommissionRate
			newCommission.UpdateTime = ctx.BlockTime()
		}
		if err := newCommission.Validate(); err != nil {
			return 0, fmt.Errorf("validator %s: %w", validator.OperatorAddress, err)
		}

		valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		if err != nil {
			return 0, err
		}
		// Same order as MsgEditValidator: the hook runs before the validator is written.
		if err := stakingKeeper.Hooks().BeforeValidatorModified(ctx, valAddr); err != nil {
			return 0, err
		}
		validator.Commission = newCommission
		if err := stakingKeeper.SetValidator(ctx, validator); err != nil {
			return 0, err
		}

		ctx.EventManager().EmitEvent(sdk.NewEvent(
			EventTypeMinCommissionRaised,
			sdk.NewAttribute(AttributeKeyValidator, validator.OperatorAddress),
			sdk.NewAttribute(AttributeKeyOldRate, oldCommission.Rate.String()),
			sdk.NewAttribute(AttributeKeyNewRate, newCommission.Rate.String()),
			sdk.NewAttribute(AttributeKeyOldMaxRate, oldCommission.MaxRate.String()),
			sdk.NewAttribute(AttributeKeyNewMaxRate, newCommission.MaxRate.String()),
		))
		modified++
	}
	return modified, nil
}
