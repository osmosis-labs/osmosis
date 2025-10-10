package v7

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/osmosis-labs/osmosis/v31/app/upgrades"
	superfluidtypes "github.com/osmosis-labs/osmosis/v31/x/superfluid/types"

	store "cosmossdk.io/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v7 upgrade.
const UpgradeName = "v7"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{wasmtypes.ModuleName, superfluidtypes.ModuleName},
	},
}
