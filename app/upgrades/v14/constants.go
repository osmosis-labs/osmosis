package v14

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/osmosis-labs/osmosis/v13/app/upgrades"
	protorevtypes "github.com/osmosis-labs/osmosis/v13/x/protorev/types"
	valsetpreftypes "github.com/osmosis-labs/osmosis/v13/x/valset-pref/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v14 upgrade.
const UpgradeName = "v14"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{valsetpreftypes.StoreKey, protorevtypes.StoreKey},
		Deleted: []string{},
	},
}
