package v16

import (
	"github.com/osmosis-labs/osmosis/v15/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"

	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v16 upgrade.
const UpgradeName = "v16"

// new token factory parameters
//
// at the current gas price of 0.0025uosmo, this corresponds to 0.1 OSMO per
// denom creation.
const NewDenomCreationGasConsume uint64 = 40_000_000

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{cltypes.StoreKey},
		Deleted: []string{},
	},
}
