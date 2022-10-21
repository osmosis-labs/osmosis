package v13

import (
	"github.com/osmosis-labs/osmosis/v12/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"

	concentratedliquiditytypes "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v9 upgrade.
const UpgradeName = "v13"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{swaproutertypes.StoreKey, concentratedliquiditytypes.StoreKey},
	},
}
