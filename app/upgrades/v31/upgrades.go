package v31

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v30/app/keepers"
	"github.com/osmosis-labs/osmosis/v30/app/upgrades"
	mintkeeper "github.com/osmosis-labs/osmosis/v30/x/mint/keeper"
	poolmanager "github.com/osmosis-labs/osmosis/v30/x/poolmanager"
	txfeestypes "github.com/osmosis-labs/osmosis/v30/x/txfees/types"
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

		err = updateTakerFeeDistribution(sdkCtx, keepers.PoolManagerKeeper, keepers.AccountKeeper)
		if err != nil {
			return nil, err
		}

		err = initializeRestrictedAddresses(sdkCtx, keepers.MintKeeper)
		if err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

// updateTakerFeeDistribution updates the community_pool and burn values in the osmo_taker_fee_distribution
// This changes taker fees from being sent to the community pool to being burned instead.
func updateTakerFeeDistribution(ctx sdk.Context, poolManagerKeeper *poolmanager.Keeper, accountKeeper *authkeeper.AccountKeeper) error {
	poolManagerParams := poolManagerKeeper.GetParams(ctx)

	// Set community_pool to 0, burn and staking rewards to 70%:30%
	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.CommunityPool = osmomath.ZeroDec()
	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.Burn = osmomath.MustNewDecFromStr("0.7")
	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.StakingRewards = osmomath.MustNewDecFromStr("0.3")

	// Set non-OSMO taker fee distribution: staking_rewards=22.5%, burn=52.5%, community_pool=25%
	poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.StakingRewards = osmomath.MustNewDecFromStr("0.225")
	poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.Burn = osmomath.MustNewDecFromStr("0.525")
	poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool = osmomath.MustNewDecFromStr("0.25")

	poolManagerKeeper.SetParams(ctx, poolManagerParams)

	// Ensure new module account exists for non‑native taker fee burn bucket. Error if it already exists.
	err := osmoutils.CreateModuleAccountByName(ctx, accountKeeper, txfeestypes.TakerFeeBurnName)
	if err != nil {
		return err
	}
	return nil
}

// initializeRestrictedAddresses sets the initial restricted asset addresses in mint params.
// These addresses represent foundation and investor wallets whose balance and staked amounts
// should be excluded from circulating supply calculations.
func initializeRestrictedAddresses(ctx sdk.Context, mintKeeper *mintkeeper.Keeper) error {
	params := mintKeeper.GetParams(ctx)

	// Initialize with the example address provided
	// This can be updated via governance in the future
	params.RestrictedAssetAddresses = []string{
		"osmo1ugku28hwyexpljrrmtet05nd6kjlrvr9jz6z00", // Example foundation/investor address
	}

	mintKeeper.SetParams(ctx, params)
	return nil
}
