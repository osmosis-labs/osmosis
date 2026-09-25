package v32

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v31/app/keepers"
	"github.com/osmosis-labs/osmosis/v31/app/upgrades"
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
		if err := seizeFrozenAllBTC(sdkCtx, keepers); err != nil {
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
