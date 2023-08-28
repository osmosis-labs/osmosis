package v14

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	ibchookstypes "github.com/osmosis-labs/osmosis/x/ibc-hooks/types"

	"github.com/osmosis-labs/osmosis/v19/app/upgrades"
	downtimetypes "github.com/osmosis-labs/osmosis/v19/x/downtime-detector/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v14 upgrade.
const UpgradeName = "v14"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{downtimetypes.StoreKey, ibchookstypes.StoreKey},
		Deleted: []string{},
	},
}
