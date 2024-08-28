package v12

import (
	"github.com/osmosis-labs/osmosis/v26/app/upgrades"
	twaptypes "github.com/osmosis-labs/osmosis/v26/x/twap/types"

	store "cosmossdk.io/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v12 upgrade.
const UpgradeName = "v12"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{twaptypes.StoreKey},
		Deleted: []string{}, // double check bech32ibc
	},
}
