package v25

import (
	"github.com/osmosis-labs/osmosis/v24/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"

	smartaccounttypes "github.com/osmosis-labs/osmosis/v24/x/smart-account/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v25 upgrade.
const UpgradeName = "v25"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			smartaccounttypes.StoreKey,
		},
		Deleted: []string{},
	},
}

var AstroportPoolIds = [...]uint64{1564, 1567, 1568, 1569, 1570, 1572, 1579, 1616, 1617, 1635, 1654}
