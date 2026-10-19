package v32

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v32/app/keepers"
	"github.com/osmosis-labs/osmosis/v32/app/upgrades"
	protorevtypes "github.com/osmosis-labs/osmosis/v32/x/protorev/types"
)

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
			if err := seizeFrozenAllBTC(sdkCtx, keepers); err != nil {
				return nil, err
			}
		} else {
			sdkCtx.Logger().Info("skipping proposal 1044 allBTC seizure on non-mainnet chain", "chain_id", sdkCtx.ChainID())
		}

		if err := moveProtorevNonBaseDenomsToSubDAO(sdkCtx, keepers); err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

// seizeFrozenAllBTC moves the proposal-authorized allBTC amount to the Liquidity subDAO.
// Proposal 1044 authorizes that denom only. The account's OSMO stays where it is.
// A balance below the authorized amount aborts the upgrade. Any surplus stays on the
// exploiter: that address can still receive allBTC, so requiring the historical balance
// exactly would let a one-unit deposit halt the upgrade block.
func seizeFrozenAllBTC(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	from, err := sdk.AccAddressFromBech32(ExploiterAddress)
	if err != nil {
		return fmt.Errorf("seizing frozen allBTC: exploiter address: %w", err)
	}
	to, err := sdk.AccAddressFromBech32(LiquiditySubDAOAddress)
	if err != nil {
		return fmt.Errorf("seizing frozen allBTC: liquidity subDAO address: %w", err)
	}

	authorized, ok := osmomath.NewIntFromString(ExpectedExploiterAllBTCAmount)
	if !ok {
		return fmt.Errorf("seizing frozen allBTC: expected amount %s is not an integer", ExpectedExploiterAllBTCAmount)
	}

	balance := keepers.BankKeeper.GetBalance(ctx, from, AllBTCDenom)
	if balance.Amount.LT(authorized) {
		return fmt.Errorf("seizing frozen allBTC: balance %s is below authorized %s %s", balance.Amount.String(), ExpectedExploiterAllBTCAmount, AllBTCDenom)
	}

	coin := sdk.NewCoin(AllBTCDenom, authorized)
	if err := keepers.BankKeeper.SendCoins(ctx, from, to, sdk.NewCoins(coin)); err != nil {
		return fmt.Errorf("seizing frozen allBTC: %w", err)
	}

	ctx.Logger().Info("seized frozen allBTC", "from", ExploiterAddress, "to", LiquiditySubDAOAddress, "amount", coin.String())
	return nil
}

// moveProtorevNonBaseDenomsToSubDAO sends every protorev module balance that is not a
// protorev base denom to the liquidity sub-DAO. Base denoms are left in place because
// the daily epoch hook already distributes them (developer fee, burn, community pool);
// everything else accumulates in the module account with no way out.
//
// Each denom is sent in its own cache context so a single denom that refuses to move
// (e.g. a tokenfactory before-send hook that errors) is logged and skipped rather than
// halting the upgrade.
func moveProtorevNonBaseDenomsToSubDAO(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	recipient, err := sdk.AccAddressFromBech32(LiquiditySubDAOAddress)
	if err != nil {
		return err
	}

	baseDenoms, err := keepers.ProtoRevKeeper.GetAllBaseDenoms(ctx)
	if err != nil {
		return err
	}
	isBaseDenom := make(map[string]bool, len(baseDenoms))
	for _, baseDenom := range baseDenoms {
		isBaseDenom[baseDenom.Denom] = true
	}

	moduleAddr := keepers.AccountKeeper.GetModuleAddress(protorevtypes.ModuleName)
	balances := keepers.BankKeeper.GetAllBalances(ctx, moduleAddr)

	moved := sdk.NewCoins()
	for _, coin := range balances {
		if isBaseDenom[coin.Denom] {
			continue
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			return keepers.BankKeeper.SendCoinsFromModuleToAccount(cacheCtx, protorevtypes.ModuleName, recipient, sdk.NewCoins(coin))
		})
		if err != nil {
			ctx.Logger().Error("v32: failed to move protorev balance to liquidity sub-DAO", "coin", coin.String(), "error", err)
			continue
		}
		moved = moved.Add(coin)
	}

	ctx.Logger().Info("v32: moved protorev non-base-denom balances to liquidity sub-DAO",
		"recipient", LiquiditySubDAOAddress, "denoms_moved", len(moved), "coins", moved.String())
	return nil
}
