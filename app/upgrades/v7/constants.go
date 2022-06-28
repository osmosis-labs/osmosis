package v7

import (
	"github.com/CosmWasm/wasmd/x/wasm"

	"github.com/osmosis-labs/osmosis/v7/app/upgrades"
	v7constants "github.com/osmosis-labs/osmosis/v7/app/upgrades/v7/constants"
	superfluidtypes "github.com/osmosis-labs/osmosis/v7/x/superfluid/types"

	store "github.com/cosmos/cosmos-sdk/store/types"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          v7constants.UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{wasm.ModuleName, superfluidtypes.ModuleName},
	},
}
