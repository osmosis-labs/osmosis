package v13

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/osmosis-labs/osmosis/v13/app/upgrades"
	concentratedliquidtytypes "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
	valsetpreftypes "github.com/osmosis-labs/osmosis/v13/x/valset-pref/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v13 upgrade.
const UpgradeName = "v13"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{valsetpreftypes.StoreKey, swaproutertypes.StoreKey, concentratedliquidtytypes.StoreKey},
		Deleted: []string{}, // double check bech32ibc
	},
}
