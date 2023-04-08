package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v15/app/keepers"
	"github.com/osmosis-labs/osmosis/v15/app/upgrades"
)

const (
	// TODO: change this to OSMO / DAI pool ID
	// https://app.osmosis.zone/pool/674
	cfmmPoolIdToLink = uint64(1)
	// TODO: make sure this is what we desire.
	desiredDenom0 = "uosmo"
	// TODO: confirm pre-launch.
	tickSpacing = 1

	// TODO: confirm that concentrated pool swap fee should equal balancer swap fee.
)

var (
	// TODO: confirm pre-launch.
	exponentAtPriceOne = sdk.OneInt().Neg()
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

		if err := createCanonicalConcentratedLiquidityPoolAndMigrationLink(ctx, cfmmPoolIdToLink, desiredDenom0, keepers); err != nil {
			return nil, err
		}

		return migrations, nil
	}
}
