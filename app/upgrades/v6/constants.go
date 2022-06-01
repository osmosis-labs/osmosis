package v6

import "github.com/osmosis-labs/osmosis/v9/app/upgrades"

const (
	// UpgradeName defines the on-chain upgrade name for the Osmosis v6 upgrade.
	UpgradeName = "v6"

	// UpgradeHeight defines the block height at which the Osmosis v6 upgrade is
	// triggered.
	UpgradeHeight = 2_464_000
)

var Fork = upgrades.Fork{
	UpgradeName:    UpgradeName,
	UpgradeHeight:  UpgradeHeight,
	BeginForkLogic: RunForkLogic,
}
