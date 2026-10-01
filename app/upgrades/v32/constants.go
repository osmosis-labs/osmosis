package v32

import (
	"github.com/osmosis-labs/osmosis/v31/app/upgrades"

	store "cosmossdk.io/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v32 upgrade.
const UpgradeName = "v32"

// Proposal 1044 seizes the allBTC frozen on the exploiter by v31.1.0 and sends
// it to the Liquidity subDAO. The amount is the base-unit balance queried on
// 2026-09-24 (22.65060846 allBTC). The proposal text rounds this to 22.650608.
// The account's only other coin is 703055696 uosmo, which this upgrade leaves
// in place.
const (
	ExploiterAddress              = "osmo1wq76r2mhqsa9yaygghuwyq4wy6dcsgf8vtzltn"
	LiquiditySubDAOAddress        = "osmo1rvq5cq2j35k7sqqz49e5e8zezl45fcywcawazh46qnc0g96d0d6sasqsgc"
	AllBTCDenom                   = "factory/osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3/alloyed/allBTC"
	ExpectedExploiterAllBTCAmount = "2265060846"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}
