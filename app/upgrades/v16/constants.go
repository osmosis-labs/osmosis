package v16

import (
	"github.com/osmosis-labs/osmosis/v15/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"

	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v16 upgrade.
const UpgradeName = "v16"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{cltypes.StoreKey, cosmwasmpooltypes.StoreKey},
		Deleted: []string{},
	},
}
