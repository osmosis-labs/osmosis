package v25

import (
	"github.com/osmosis-labs/osmosis/v24/app/upgrades"
	bridgetypes "github.com/osmosis-labs/osmosis/v24/x/bridge/types"

	store "github.com/cosmos/cosmos-sdk/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v24 upgrade.
const UpgradeName = "v25"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{bridgetypes.StoreKey},
		Deleted: []string{},
	},
}
