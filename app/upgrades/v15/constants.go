package v15

import (
	"github.com/osmosis-labs/osmosis/v14/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v15 upgrade.
const UpgradeName = "v15"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
    },
}
