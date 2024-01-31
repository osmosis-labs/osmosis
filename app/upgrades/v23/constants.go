package v23

import (
	"github.com/osmosis-labs/osmosis/v22/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/skip-mev/block-sdk/x/auction/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v22 upgrade.
const UpgradeName = "v23"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{types.ModuleName},
		Deleted: []string{},
	},
}

