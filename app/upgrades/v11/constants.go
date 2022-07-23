package v11

import (
	"github.com/osmosis-labs/osmosis/v10/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v11 upgrade.
const UpgradeName = "v11"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
