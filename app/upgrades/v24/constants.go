package v24

import (
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"

	"github.com/osmosis-labs/osmosis/v24/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v24 upgrade.
const UpgradeName = "v24"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{ibcwasmtypes.StoreKey, icacontrollertypes.StoreKey},
		Deleted: []string{},
	},
}

// TODO: Change to pool IDs that we actually want to migrate
var MigratedSpreadFactorAccumulatorPoolIDs = map[uint64]struct{}{
	1423: {},
	1213: {},
	1298: {},
	1297: {},
	1292: {},
	1431: {},
}
