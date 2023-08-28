package v4

import (
	"github.com/osmosis-labs/osmosis/v19/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v4 upgrade.
const UpgradeName = "v4"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
