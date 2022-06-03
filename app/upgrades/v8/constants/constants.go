// package v8constants contains constants related to the v8 upgrade.
// It is in its own package to eliminate import cycle issues.
package v8constants

const (
	// UpgradeName defines the on-chain upgrade name for the Osmosis v8 upgrade.
	UpgradeName = "v8"

	// UpgradeHeight defines the block height at which the Osmosis v8 upgrade is
	// triggered.
	// Block height 4_402_000, approximately 4PM UTC, May 15th
	UpgradeHeight = 4402000
)
