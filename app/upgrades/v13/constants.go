package v13

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/osmosis-labs/osmosis/v12/app/upgrades"
	valsetpreftypes "github.com/osmosis-labs/osmosis/v12/x/valset-pref/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v13 upgrade.
const UpgradeName = "v13"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{valsetpreftypes.StoreKey},
		Deleted: []string{}, // double check bech32ibc
	},
}
