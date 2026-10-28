package v33

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v32/app/keepers"
)

// CollectHolders exposes collectHolders so tests can exercise the holder cap without creating
// MaxWBTCHoldersToMigrate accounts.
var CollectHolders = collectHolders

// ParseAddresses, WBTCUnindexedHolders and NobleUSDCUnindexedHolders expose the embedded lists.
var (
	ParseAddresses            = parseAddresses
	WBTCUnindexedHolders      = wbtcUnindexedHolders
	NobleUSDCUnindexedHolders = nobleUSDCUnindexedHolders
)

// SetAlloyMigrationBuildersForTest replaces the osmosis-1 migrations for one test. The returned
// function restores the production builders.
func SetAlloyMigrationBuildersForTest(builders []AlloyMigrationBuilder) func() {
	prev := alloyMigrationBuilders
	alloyMigrationBuilders = builders
	return func() { alloyMigrationBuilders = prev }
}

// MainnetNobleUSDCMigration exposes the production Noble USDC config builder.
func MainnetNobleUSDCMigration(ctx sdk.Context, keepers *keepers.AppKeepers) (AlloyMigration, error) {
	return mainnetNobleUSDCMigration(ctx, keepers)
}
