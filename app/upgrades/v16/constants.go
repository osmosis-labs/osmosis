package v16

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/osmosis-labs/osmosis/v15/app/upgrades"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v15 upgrade.
const UpgradeName = "v16"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{cltypes.StoreKey},
		Deleted: []string{},
	},
}
