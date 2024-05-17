package v26

import (
	"github.com/osmosis-labs/osmosis/v25/app/upgrades"

	storetypes "cosmossdk.io/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v26 upgrade.
const UpgradeName = "v26"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}
