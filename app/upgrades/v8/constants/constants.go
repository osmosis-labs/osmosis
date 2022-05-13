// package v8constants contains constants related to the v8 upgrade.
// It is in its own package to eliminate import cycle issues.
package v8constants

const (
	// UpgradeName defines the on-chain upgrade name for the Osmosis v3 upgrade.
	UpgradeName = "v8"

	// UpgradeHeight defines the block height at which the Osmosis v8 upgrade is
	// triggered.
	// TODO: Choose upgrade height
	UpgradeHeight = 100_000
)
