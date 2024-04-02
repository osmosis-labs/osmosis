package v23

import (
	"github.com/osmosis-labs/osmosis/v24/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v23 upgrade.
const UpgradeName = "v23"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}

var MigratedIncentiveAccumulatorPoolIDs = map[uint64]struct{}{
	1423: {},
	1213: {},
	1298: {},
	1297: {},
	1292: {},
	1431: {},
}
