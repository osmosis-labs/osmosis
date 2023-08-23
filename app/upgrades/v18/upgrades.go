package v18

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v17/app/keepers"
	"github.com/osmosis-labs/osmosis/v17/app/upgrades"
	lockuptypes "github.com/osmosis-labs/osmosis/v17/x/lockup/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v17/x/protorev/types"
)

var (
	OSMO        = "uosmo"
	AKTIBCDenom = "ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4"
	pool3Denom  = "gamm/pool/3"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		addr, err := sdk.AccAddressFromBech32("osmo1urn0pnx8fl5kt89r5nzqd8htruq7skadc2xdk3")
		if err != nil {
			return nil, err
		}

		err = keepers.BankKeeper.MintCoins(ctx, protorevtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(OSMO, sdk.NewInt(50000000000))))
		if err != nil {
			return nil, err
		}
		err = keepers.BankKeeper.SendCoinsFromModuleToAccount(ctx, protorevtypes.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(OSMO, sdk.NewInt(50000000000))))
		if err != nil {
			return nil, err
		}

		aktGAMMPool, err := keepers.GAMMKeeper.GetPool(ctx, 3)
		if err != nil {
			return nil, err
		}

		sharesOut, err := keepers.GAMMKeeper.JoinSwapExactAmountIn(ctx, addr, aktGAMMPool.GetId(), sdk.NewCoins(sdk.NewCoin(OSMO, sdk.NewInt(50000000000))), sdk.ZeroInt())
		if err != nil {
			return nil, err
		}
		aktSharesDenom := fmt.Sprintf("gamm/pool/%d", aktGAMMPool.GetId())
		shareCoins := sdk.NewCoins(sdk.NewCoin(aktSharesDenom, sharesOut))
		lock, err := keepers.LockupKeeper.CreateLock(ctx, addr, shareCoins, time.Hour*24*7)
		if err != nil {
			return nil, err
		}

		value := keepers.LockupKeeper.GetPeriodLocksAccumulation(ctx, lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "gamm/pool/3",
			Duration:      time.Hour * 24 * 7,
		})
		ctx.Logger().Info("VALUE PRE: ", value)

		// Clear gamm/pool/3 denom accumulation store
		keepers.LockupKeeper.ClearDenomAccumulationStore(ctx, pool3Denom)

		// Remove the lockup created for pool 3 above
		err = keepers.LockupKeeper.ForceUnlock(ctx, lock)
		if err != nil {
			return nil, err
		}

		value = keepers.LockupKeeper.GetPeriodLocksAccumulation(ctx, lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "gamm/pool/3",
			Duration:      time.Hour * 24 * 7,
		})
		ctx.Logger().Info("VALUE POST: ", value)

		return migrations, nil
	}
}
