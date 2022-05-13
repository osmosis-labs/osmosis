package v8

import "github.com/osmosis-labs/osmosis/v7/app/upgrades"

const (
	// UpgradeName defines the on-chain upgrade name for the Osmosis v3 upgrade.
	UpgradeName = "v8"

	// UpgradeHeight defines the block height at which the Osmosis v3 upgrade is
	// triggered.
	// TODO: Choose upgrade height
	UpgradeHeight = 100_000
)

var Fork = upgrades.Fork{
	UpgradeName:    UpgradeName,
	UpgradeHeight:  UpgradeHeight,
	BeginForkLogic: RunForkLogic,
}
