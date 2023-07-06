package v7

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/osmosis-labs/osmosis/v16/app/upgrades"
	superfluidtypes "github.com/osmosis-labs/osmosis/v16/x/superfluid/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v7 upgrade.
const UpgradeName = "v7"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{wasm.ModuleName, superfluidtypes.ModuleName},
	},
}
