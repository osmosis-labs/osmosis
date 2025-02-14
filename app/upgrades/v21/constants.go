package v21

import (
	"github.com/osmosis-labs/osmosis/v29/app/upgrades"

	store "cosmossdk.io/store/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v21 upgrade.
const (
	UpgradeName    = "v21"
	TestingChainId = "testing-chain-id"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			// v47 modules
			crisistypes.ModuleName,
			consensustypes.ModuleName,
		},
		Deleted: []string{},
	},
}
