package v19

import (
	"github.com/osmosis-labs/osmosis/v17/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v17 upgrade.
const UpgradeName = "v19"

var accum_stores_to_fix = []int{3, 5, 7, 9, 15, 497}

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}
