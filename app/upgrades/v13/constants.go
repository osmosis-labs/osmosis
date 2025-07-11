package v13

import (
	store "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/v30/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v13 upgrade.
const UpgradeName = "v13"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}
